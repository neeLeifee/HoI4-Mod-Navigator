package main

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

// FOLDERID_Documents respects relocated Documents and localized Explorer names.
var documentsID = syscall.GUID{Data1: 0xfdd39ad0, Data2: 0x238f, Data3: 0x46af, Data4: [8]byte{0xad, 0xb4, 0x6c, 0x85, 0x48, 0x03, 0x69, 0xc7}}

func localModsDirectory() string {
	var p *uint16
	hr, _, _ := shell.NewProc("SHGetKnownFolderPath").Call(uintptr(unsafe.Pointer(&documentsID)), 0x4000, 0, uintptr(unsafe.Pointer(&p))) // KF_FLAG_DONT_VERIFY
	if p != nil {
		defer ole.NewProc("CoTaskMemFree").Call(uintptr(unsafe.Pointer(p)))
	}
	if failed(hr) || p == nil {
		return ""
	}
	for i := 0; i < 32768; i++ {
		if *(*uint16)(unsafe.Add(unsafe.Pointer(p), i*2)) == 0 {
			root := filepath.Join(syscall.UTF16ToString(unsafe.Slice(p, i)), "Paradox Interactive", "Hearts of Iron IV", "mod")
			if localPath(root) == nil {
				return root
			}
			return ""
		}
	}
	return ""
}
