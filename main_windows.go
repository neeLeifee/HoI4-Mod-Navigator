package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

const appName = "HoI4 Mod Navigator"

var kernel = syscall.NewLazyDLL("kernel32.dll")
var user = syscall.NewLazyDLL("user32.dll")
var shell = syscall.NewLazyDLL("shell32.dll")
var ole = syscall.NewLazyDLL("ole32.dll")
var adv = syscall.NewLazyDLL("advapi32.dll")
var moveEx = kernel.NewProc("MoveFileExW")

func wide(s string) *uint16 {
	p, e := syscall.UTF16PtrFromString(s)
	if e != nil {
		panic(e)
	}
	return p
}
func ptr(p *uint16) uintptr { return uintptr(unsafe.Pointer(p)) }
func callErr(e error) error {
	if e == syscall.Errno(0) {
		return errors.New("Ошибка Windows API")
	}
	return e
}
func appDir() string       { return filepath.Join(os.Getenv("LOCALAPPDATA"), "HoI4ModNavigator") }
func installedExe() string { return filepath.Join(appDir(), "HoI4ModNavigator.exe") }
func configPath() string   { return filepath.Join(appDir(), "settings.json") }
func loadConfig() (Config, error) {
	c := defaults()
	c.LocalModsRoot = localModsDirectory()
	b, e := os.ReadFile(configPath())
	if os.IsNotExist(e) {
		return c, nil
	}
	if e != nil {
		return c, e
	}
	c.MenuVersion = 0 // distinguish old files even when defaults change
	if e = json.Unmarshal(b, &c); e != nil {
		return c, e
	}
	if c.MenuVersion < 13 {
		c.Filter = true
	}
	if c.MenuVersion < 20 {
		c.LegacyMenu = false
		c.MenuVersion = 20
	}
	c.MenuVersion = 21
	return c, nil
}
func atomicBytes(p string, b []byte) error {
	f, e := os.CreateTemp(filepath.Dir(p), ".navigator-*")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return replace(f.Name(), p, true)
}
func replace(from, to string, overwrite bool) error {
	var flag uintptr = 8
	if overwrite {
		flag |= 1
	}
	r, _, e := moveEx.Call(ptr(wide(from)), ptr(wide(to)), flag)
	if r == 0 {
		return callErr(e)
	}
	return nil
}
func saveConfig(c Config) error {
	if e := os.MkdirAll(appDir(), 0700); e != nil {
		return e
	}
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return e
	}
	return atomicBytes(configPath(), b)
}
func installSelf() error {
	src, e := os.Executable()
	if e != nil {
		return e
	}
	dst := installedExe()
	if strings.EqualFold(filepath.Clean(src), filepath.Clean(dst)) {
		return nil
	}
	if e = os.MkdirAll(appDir(), 0700); e != nil {
		return e
	}
	b, e := os.ReadFile(src)
	if e != nil {
		return e
	}
	return atomicBytes(dst, b)
}
func localPath(p string) error {
	if strings.ContainsAny(p, "\x00\r\n") || !filepath.IsAbs(p) || len(filepath.VolumeName(p)) != 2 || filepath.VolumeName(p)[1] != ':' || strings.Contains(p[2:], ":") {
		return errors.New("Поддерживаются обычные абсолютные пути на локальном диске (например D:\\Mods)")
	}
	return nil
}

// Reject junctions, symlinks and other reparse points in every existing component.
func checkChain(p string, allowMissing bool) error {
	if e := localPath(p); e != nil {
		return e
	}
	p = filepath.Clean(p)
	vol := filepath.VolumeName(p)
	cur := vol + `\`
	tail := strings.TrimPrefix(p, cur)
	for _, part := range strings.Split(tail, `\`) {
		if part == "" {
			continue
		}
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return errors.New("Недопустимое окончание имени папки или файла")
		}
		cur = filepath.Join(cur, part)
		a, e := syscall.GetFileAttributes(wide(cur))
		if e != nil {
			if allowMissing && (e == syscall.ERROR_FILE_NOT_FOUND || e == syscall.ERROR_PATH_NOT_FOUND) {
				continue
			}
			return fmt.Errorf("%s: %w", cur, e)
		}
		if a&0x400 != 0 {
			return fmt.Errorf("Ссылки/junction/reparse points не поддерживаются: %s", cur)
		}
	}
	return nil
}
func rootCheck(p string) error {
	if e := localPath(p); e != nil {
		return e
	}
	if filepath.Clean(p) == filepath.VolumeName(p)+`\` {
		return errors.New("Нельзя выбрать корень диска")
	}
	if e := checkChain(p, false); e != nil {
		return e
	}
	if e := existingDir(p); e != nil {
		return e
	}
	real, e := filepath.EvalSymlinks(p)
	if e != nil {
		return e
	}
	if !strings.EqualFold(filepath.Clean(real), filepath.Clean(p)) {
		return errors.New("Выбери полный обычный путь, без псевдонимов и ссылок")
	}
	return nil
}
func validate(c Config) error {
	if e := validateConfig(c, rootCheck); e != nil {
		return e
	}
	st, e := os.Stat(filepath.Join(c.Game, "hoi4.exe"))
	if e != nil || !st.Mode().IsRegular() {
		return errors.New("В папке оригинальной игры не найден hoi4.exe")
	}
	return nil
}
func regular(p string) error {
	if e := checkChain(p, false); e != nil {
		return e
	}
	st, e := os.Stat(p)
	if e != nil {
		return e
	}
	if !st.Mode().IsRegular() {
		return errors.New("Ожидался обычный файл")
	}
	return nil
}
func ensureDir(root, p string, create bool) error {
	if !inside(root, p) {
		return errors.New("Путь назначения вне мода")
	}
	if e := checkChain(p, create); e != nil {
		return e
	}
	if create {
		if e := os.MkdirAll(p, 0755); e != nil {
			return e
		}
	}
	if e := checkChain(p, false); e != nil {
		return e
	}
	return existingDir(p)
}
func explorer(p string, selectFile bool) error {
	win := os.Getenv("WINDIR")
	if win == "" {
		win = `C:\Windows`
	}
	exe := filepath.Join(win, "explorer.exe")
	cmd := exec.Command(exe)
	// Explorer uses comma-separated switches rather than standard argv parsing.
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: explorerCommand(exe, p, selectFile)}
	return cmd.Start()
}
func vscode() (string, error) {
	for _, p := range []string{filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Microsoft VS Code", "Code.exe"), filepath.Join(os.Getenv("ProgramFiles"), "Microsoft VS Code", "Code.exe"), filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft VS Code", "Code.exe")} {
		if filepath.IsAbs(p) {
			if s, e := os.Stat(p); e == nil && s.Mode().IsRegular() {
				return p, nil
			}
		}
	}
	if p, e := exec.LookPath("Code.exe"); e == nil {
		return p, nil
	}
	return "", errors.New("VS Code не найден. Установи Visual Studio Code или добавь Code.exe в PATH")
}
func copyFile(src, dst, root string, create bool) error {
	if e := regular(src); e != nil {
		return e
	}
	if !inside(root, dst) {
		return errors.New("Путь назначения вне мода")
	}
	if e := ensureDir(root, filepath.Dir(dst), create); e != nil {
		return e
	}
	if e := checkChain(dst, true); e != nil {
		return e
	}
	overwrite := false
	if st, e := os.Stat(dst); e == nil {
		if !st.Mode().IsRegular() {
			return errors.New("На месте файла находится папка или специальный объект")
		}
		if !ask("Файл уже существует:\n" + dst + "\n\nПерезаписать его оригиналом?") {
			return nil
		}
		overwrite = true
	} else if !os.IsNotExist(e) {
		return e
	}
	in, e := os.Open(src)
	if e != nil {
		return e
	}
	defer in.Close()
	out, e := os.CreateTemp(filepath.Dir(dst), ".hoi4nav-*")
	if e != nil {
		return e
	}
	defer os.Remove(out.Name())
	if _, e = io.Copy(out, in); e != nil {
		out.Close()
		return e
	}
	if e = out.Sync(); e != nil {
		out.Close()
		return e
	}
	if e = out.Close(); e != nil {
		return e
	}
	if e = checkChain(dst, true); e != nil {
		return e
	}
	if e = replace(out.Name(), dst, overwrite); e != nil {
		return e
	}
	return explorer(dst, true)
}
func runAction(action, id, p string) error {
	c, e := loadConfig()
	if e != nil {
		return e
	}
	if e = validate(c); e != nil {
		return e
	}
	if e = localPath(p); e != nil {
		return e
	}
	p = filepath.Clean(p)
	if action == "auto-folder" && id == "auto" {
		if e = checkChain(p, false); e != nil {
			return e
		}
		if e = existingDir(p); e != nil {
			return e
		}
		choices := autoChoices(c, p)
		if len(choices) == 0 {
			return errors.New("Эта папка находится вне настроенной игры/модов или переход в этом направлении выключен")
		}
		i := 0
		if len(choices) > 1 {
			i = chooseDestination(choices)
		}
		if i < 0 {
			return nil
		}
		// Use the same guarded implementation as the direct verbs.
		return runAction(choices[i].action, choices[i].id, p)
	}
	// Auto-discovered mods are read-only sources, never write destinations.
	if id == "local" && (action == "reverse-local" || action == "reverse-local-file" || action == "compare-local") {
		root, err := localModRoot(c, p)
		if err != nil {
			return err
		}
		if err = rootCheck(root); err != nil {
			return err
		}
		dst, err := mapped(root, c.Game, p)
		if err != nil {
			return err
		}
		if action == "compare-local" {
			if !c.Compare {
				return errors.New("Функция выключена")
			}
			if err = regular(p); err != nil {
				return err
			}
			if err = regular(dst); err != nil {
				return fmt.Errorf("Оригинал отсутствует или недоступен: %w", err)
			}
			code, err := vscode()
			if err != nil {
				return err
			}
			return exec.Command(code, "--new-window", "--diff", dst, p).Start()
		}
		if !c.Reverse {
			return errors.New("Функция выключена")
		}
		if action == "reverse-local-file" {
			if err = regular(p); err != nil {
				return err
			}
			if err = regular(dst); err != nil {
				return fmt.Errorf("В ванильной игре нет соответствующего файла: %w", err)
			}
			return explorer(dst, true)
		}
		if err = checkChain(p, false); err != nil {
			return err
		}
		if err = existingDir(p); err != nil {
			return err
		}
		if err = ensureDir(c.Game, dst, false); err != nil {
			return fmt.Errorf("В оригинальной игре такой папки нет или она недоступна: %w", err)
		}
		return explorer(dst, false)
	}
	var m Mod
	for _, x := range c.Mods {
		if x.ID == id {
			m = x
			break
		}
	}
	if m.ID == "" {
		return errors.New("Мод не найден. Примени настройки заново")
	}
	switch action {
	case "open", "parent", "copy":
		if (action == "copy" && !c.Copy) || (action != "copy" && !c.Open) {
			return errors.New("Функция выключена")
		}
		dst, e := mapped(c.Game, m.Path, p)
		if e != nil {
			return errors.New("Открой папку или файл оригинальной HoI4")
		}
		if e = checkChain(p, false); e != nil {
			return e
		}
		if action == "copy" {
			return copyFile(p, dst, m.Path, c.Create)
		}
		if action == "parent" {
			if e = regular(p); e != nil {
				return e
			}
			dst = filepath.Dir(dst)
		} else if e = existingDir(p); e != nil {
			return e
		}
		if e = ensureDir(m.Path, dst, c.Create); e != nil {
			return e
		}
		return explorer(dst, false)
	case "reverse", "reverse-file":
		if !c.Reverse {
			return errors.New("Функция выключена")
		}
		dst, e := mapped(m.Path, c.Game, p)
		if e != nil {
			return errors.New("Открой папку выбранного мода")
		}
		if e = checkChain(p, false); e != nil {
			return e
		}
		if action == "reverse-file" {
			if e = regular(p); e != nil {
				return e
			}
			if e = regular(dst); e != nil {
				return fmt.Errorf("В ванильной игре нет соответствующего файла или он недоступен:\n%s\n\n%w", dst, e)
			}
			return explorer(dst, true)
		}
		if e = existingDir(p); e != nil {
			return e
		}
		if e = ensureDir(c.Game, dst, false); e != nil {
			return fmt.Errorf("В оригинальной игре такой папки нет или она недоступна: %w", e)
		}
		return explorer(dst, false)
	case "compare":
		if !c.Compare {
			return errors.New("Функция выключена")
		}
		var orig, mod string
		if inside(c.Game, p) {
			orig = p
			mod, e = mapped(c.Game, m.Path, p)
		} else if inside(m.Path, p) {
			mod = p
			orig, e = mapped(m.Path, c.Game, p)
		} else {
			return errors.New("Файл должен находиться в оригинальной игре или выбранном моде")
		}
		if e != nil {
			return e
		}
		if e = regular(orig); e != nil {
			return fmt.Errorf("Оригинал отсутствует или недоступен: %w", e)
		}
		if e = regular(mod); e != nil {
			return fmt.Errorf("Файл мода отсутствует или недоступен: %w", e)
		}
		code, e := vscode()
		if e != nil {
			return e
		}
		return exec.Command(code, "--new-window", "--diff", orig, mod).Start()
	}
	return errors.New("Неизвестная команда")
}
func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if len(os.Args) > 1 {
		if len(os.Args) == 4 && os.Args[1] == "action" {
			message(0, "Некорректные аргументы", 0x10)
			return
		}
		if len(os.Args) == 5 && os.Args[1] == "action" {
			if e := runAction(os.Args[2], os.Args[3], os.Args[4]); e != nil {
				message(0, e.Error(), 0x30)
			}
			return
		}
		if len(os.Args) != 2 || os.Args[1] != "settings" {
			message(0, "Неизвестные аргументы", 0x10)
			return
		}
	}
	runGUI()
}
