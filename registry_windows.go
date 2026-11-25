package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var regCreate = adv.NewProc("RegCreateKeyExW")
var regSet = adv.NewProc("RegSetValueExW")
var regDelete = adv.NewProc("RegDeleteTreeW")

func setReg(path, name, value string) error {
	var h syscall.Handle
	var disp uint32
	r, _, _ := regCreate.Call(uintptr(syscall.HKEY_CURRENT_USER), ptr(wide(path)), 0, 0, 0, 0x20006|0x0100, 0, uintptr(unsafe.Pointer(&h)), uintptr(unsafe.Pointer(&disp)))
	if r != 0 {
		return syscall.Errno(r)
	}
	defer syscall.RegCloseKey(h)
	b, e := syscall.UTF16FromString(value)
	if e != nil {
		return e
	}
	var n uintptr
	if name != "" {
		n = ptr(wide(name))
	}
	r, _, _ = regSet.Call(uintptr(h), n, 0, 1, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)*2))
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

//go:embed HoI4NavigatorShell_v21.dll
var shellDLL []byte

func installDLL() (string, error) {
	p := filepath.Join(appDir(), dllName)
	if b, e := os.ReadFile(p); e == nil && bytes.Equal(b, shellDLL) {
		return p, nil
	}
	if e := atomicBytes(p, shellDLL); e != nil {
		return "", e
	}
	return p, nil
}
func removeMenus() error {
	var first error
	keys := removalKeys()
	for i := range classIDs {
		keys = append(keys, handlerKey(i), classKey(i))
	}
	keys = append(keys, shellKey)
	for _, key := range keys {
		r, _, _ := regDelete.Call(uintptr(syscall.HKEY_CURRENT_USER), ptr(wide(key)))
		if r != 0 && r != 2 && r != 3 && first == nil {
			first = syscall.Errno(r)
		}
	}
	notifyShell()
	return first
}
func notifyShell() { shell.NewProc("SHChangeNotify").Call(0x08000000, 0, 0, 0) }

func registerMenus(c Config) error {
	var entries []regEntry
	var e error
	if c.LegacyMenu {
		entries, e = menuEntries(c, installedExe())
	} else {
		var dll string
		dll, e = installDLL()
		if e == nil {
			entries, e = dynamicEntries(c, installedExe(), dll)
		}
	}
	if e != nil {
		return e
	}
	if e = removeMenus(); e != nil {
		return e
	}
	for _, v := range entries {
		if e = setReg(v.path, v.name, v.value); e != nil {
			_ = removeMenus()
			if !c.LegacyMenu {
				return fmt.Errorf("Не удалось зарегистрировать динамическое меню: %w. Включи совместимый режим и примени настройки", e)
			}
			return e
		}
	}
	notifyShell()
	return nil
}
