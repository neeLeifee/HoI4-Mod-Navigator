package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// Native Common Item Dialog. The GUI thread is initialized as STA by showWindow.
var clsidFileOpenDialog = syscall.GUID{Data1: 0xdc1c5a9c, Data2: 0xe88a, Data3: 0x4dde, Data4: [8]byte{0xa5, 0xa1, 0x60, 0xf8, 0x2a, 0x20, 0xae, 0xf7}}
var iidFileDialog = syscall.GUID{Data1: 0x42f85136, Data2: 0xdb7e, Data3: 0x439c, Data4: [8]byte{0x85, 0xf1, 0xe4, 0x07, 0x5d, 0x13, 0x5f, 0xc8}}
var iidShellItem = syscall.GUID{Data1: 0x43826d1e, Data2: 0xe718, Data3: 0x42ee, Data4: [8]byte{0xbc, 0x55, 0xa1, 0xe2, 0x61, 0xc3, 0x7b, 0xfe}}

type fileDialog struct{ vtbl *[27]uintptr }
type shellItem struct{ vtbl *[8]uintptr }

func failed(hr uintptr) bool { return int32(hr) < 0 }
func browse(title, initial string) (string, bool) {
	var d *fileDialog
	hr, _, _ := ole.NewProc("CoCreateInstance").Call(uintptr(unsafe.Pointer(&clsidFileOpenDialog)), 0, 1, uintptr(unsafe.Pointer(&iidFileDialog)), uintptr(unsafe.Pointer(&d)))
	report := func(stage string, code uintptr) (string, bool) {
		message(hwnd, fmt.Sprintf("Не удалось открыть диалог выбора папки (%s, 0x%08X).\nПуть можно ввести вручную.", stage, uint32(code)), 0x30)
		return "", false
	}
	if failed(hr) || d == nil {
		return report("CoCreateInstance", hr)
	}
	defer syscall.SyscallN(d.vtbl[2], uintptr(unsafe.Pointer(d)))
	var opts uint32
	hr, _, _ = syscall.SyscallN(d.vtbl[10], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&opts)))
	if failed(hr) {
		return report("GetOptions", hr)
	}
	// FOS_PICKFOLDERS | FOS_FORCEFILESYSTEM | FOS_PATHMUSTEXIST | FOS_NOCHANGEDIR
	opts |= 0x20 | 0x40 | 0x800 | 0x8
	hr, _, _ = syscall.SyscallN(d.vtbl[9], uintptr(unsafe.Pointer(d)), uintptr(opts))
	if failed(hr) {
		return report("SetOptions", hr)
	}
	syscall.SyscallN(d.vtbl[17], uintptr(unsafe.Pointer(d)), ptr(wide(title)))
	syscall.SyscallN(d.vtbl[18], uintptr(unsafe.Pointer(d)), ptr(wide("Выбрать папку")))
	initial = strings.TrimSpace(initial)
	if s, e := os.Stat(initial); e == nil && s.IsDir() {
		var item *shellItem
		hr, _, _ = shell.NewProc("SHCreateItemFromParsingName").Call(ptr(wide(initial)), 0, uintptr(unsafe.Pointer(&iidShellItem)), uintptr(unsafe.Pointer(&item)))
		if !failed(hr) && item != nil {
			syscall.SyscallN(d.vtbl[12], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(item)))
			syscall.SyscallN(item.vtbl[2], uintptr(unsafe.Pointer(item)))
		}
	}
	hr, _, _ = syscall.SyscallN(d.vtbl[3], uintptr(unsafe.Pointer(d)), hwnd)
	if uint32(hr) == 0x800704c7 {
		return "", false
	}
	if failed(hr) {
		return report("Show", hr)
	}
	var result *shellItem
	hr, _, _ = syscall.SyscallN(d.vtbl[20], uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&result)))
	if failed(hr) || result == nil {
		return report("GetResult", hr)
	}
	defer syscall.SyscallN(result.vtbl[2], uintptr(unsafe.Pointer(result)))
	var path *uint16
	// SIGDN_FILESYSPATH
	hr, _, _ = syscall.SyscallN(result.vtbl[5], uintptr(unsafe.Pointer(result)), 0x80058000, uintptr(unsafe.Pointer(&path)))
	if failed(hr) || path == nil {
		return report("GetDisplayName", hr)
	}
	defer ole.NewProc("CoTaskMemFree").Call(uintptr(unsafe.Pointer(path)))
	chars := make([]uint16, 0, 260)
	for i := 0; i < 32768; i++ {
		ch := *(*uint16)(unsafe.Add(unsafe.Pointer(path), i*2))
		if ch == 0 {
			return syscall.UTF16ToString(chars), true
		}
		chars = append(chars, ch)
	}
	return report("Длина пути", 0x80070057)
}
