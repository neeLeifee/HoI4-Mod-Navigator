package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Mod struct {
	ID   string
	Name string
	Path string
}
type Config struct {
	Game                                         string
	Mods                                         []Mod
	Open, Copy, Reverse, Compare, Create, Filter bool
	MenuVersion                                  int
	LegacyMenu                                   bool
}

func defaults() Config {
	return Config{Open: true, Copy: true, Reverse: true, Compare: true, Create: true, Filter: true, MenuVersion: 20}
}
func newID() string {
	var b [12]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b[:])
}
func relative(root, path string) (string, error) {
	if !filepath.IsAbs(root) || !filepath.IsAbs(path) {
		return "", errors.New("Требуется абсолютный путь")
	}
	r, e := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if e != nil {
		return "", e
	}
	if r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) || filepath.IsAbs(r) {
		return "", errors.New("Путь находится вне выбранной папки")
	}
	return r, nil
}
func inside(root, path string) bool { _, e := relative(root, path); return e == nil }
func mapped(src, dst, path string) (string, error) {
	r, e := relative(src, path)
	if e != nil {
		return "", e
	}
	p := filepath.Join(dst, r)
	if !inside(dst, p) {
		return "", errors.New("Путь за пределами папки назначения")
	}
	return p, nil
}
func validateConfig(c Config, rootCheck func(string) error) error {
	if e := rootCheck(c.Game); e != nil {
		return fmt.Errorf("Папка игры: %w", e)
	}
	if len(c.Mods) == 0 {
		return errors.New("Добавь хотя бы один мод")
	}
	if len(c.Mods) > 20 {
		return errors.New("Максимум 20 модов")
	}
	ids := map[string]bool{}
	names := map[string]bool{}
	roots := []string{c.Game}
	for _, m := range c.Mods {
		b, e := hex.DecodeString(m.ID)
		if e != nil || len(b) != 12 || ids[m.ID] {
			return errors.New("Некорректный идентификатор мода")
		}
		ids[m.ID] = true
		n := strings.ToLower(strings.TrimSpace(m.Name))
		if n == "" || len([]rune(n)) > 80 || strings.ContainsAny(n, "\r\n\x00&") {
			return errors.New("Название: от 1 до 80 символов, без & и переносов строки")
		}
		if names[n] {
			return errors.New("Названия модов должны различаться")
		}
		names[n] = true
		if e := rootCheck(m.Path); e != nil {
			return fmt.Errorf("Мод %s: %w", m.Name, e)
		}
		for _, r := range roots {
			if inside(r, m.Path) || inside(m.Path, r) {
				return errors.New("Папки игры и модов не должны совпадать или находиться одна внутри другой")
			}
		}
		roots = append(roots, m.Path)
	}
	return nil
}
func existingDir(p string) error {
	st, e := os.Stat(p)
	if e != nil {
		return e
	}
	if !st.IsDir() {
		return errors.New("Это не папка")
	}
	return nil
}

type menuItem struct {
	label, action, id string
	roots             []string
}

func items(c Config, scope int) []menuItem {
	var out []menuItem
	// Keep the automatic dispatcher as the safe default.
	if scope == 0 {
		if c.Open || c.Reverse {
			label := "Открыть соответствующую папку (HoI4)"
			if len(c.Mods) == 1 {
				label = "Перейти между ваниллой и «" + c.Mods[0].Name + "»"
			}
			return []menuItem{{label, "auto-folder", "auto", nil}}
		}
		return nil
	}
	for _, m := range c.Mods {
		if scope != 2 {
			if c.Open {
				out = append(out, menuItem{"Открыть идентичную папку в «" + m.Name + "»", "open", m.ID, []string{c.Game}})
			}
			if c.Reverse {
				out = append(out, menuItem{"Открыть оригинальную папку из «" + m.Name + "»", "reverse", m.ID, []string{m.Path}})
			}
		} else {
			if c.Open {
				out = append(out, menuItem{"Открыть папку файла в «" + m.Name + "»", "parent", m.ID, []string{c.Game}})
			}
			if c.Reverse {
				out = append(out, menuItem{"Показать файл в ванильной игре (из «" + m.Name + "»)", "reverse-file", m.ID, []string{m.Path}})
			}
			if c.Copy {
				out = append(out, menuItem{"Скопировать файл в «" + m.Name + "»", "copy", m.ID, []string{c.Game}})
			}
			if c.Compare {
				out = append(out, menuItem{"Сравнить с «" + m.Name + "» / оригиналом (VS Code)", "compare", m.ID, []string{c.Game, m.Path}})
			}
		}
	}
	return out
}

// Use literal quotes around the full path; Windows file names cannot contain quotes.
func explorerCommand(exe, path string, selectFile bool) string {
	switches := `/n,`
	if selectFile {
		switches += `/select,`
	}
	return `"` + exe + `" ` + switches + `"` + path + `"`
}

// Resolve only against configured roots, not folder names or prefixes alone.
func autoChoices(c Config, path string) []menuItem {
	var out []menuItem
	if inside(c.Game, path) {
		if c.Open {
			for _, m := range c.Mods {
				out = append(out, menuItem{m.Name, "open", m.ID, []string{c.Game}})
			}
		}
		return out
	}
	if c.Reverse {
		for _, m := range c.Mods {
			if inside(m.Path, path) {
				return []menuItem{{"Ванильная игра", "reverse", m.ID, []string{m.Path}}}
			}
		}
	}
	return nil
}
