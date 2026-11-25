#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
make_lib(){ local name=$1; shift; printf 'LIBRARY %s.dll\nEXPORTS\n' "$name" > "$tmp/$name.def"; printf '%s\n' "$@" >> "$tmp/$name.def"; lld-link /lib /machine:x64 /def:"$tmp/$name.def" /out:"$tmp/$name.lib"; }
make_lib kernel32 GetProcessHeap HeapAlloc HeapFree CompareStringOrdinal GetFileAttributesW
make_lib advapi32 RegGetValueW
make_lib shell32 SHGetPathFromIDListEx DragQueryFileW ShellExecuteW
make_lib ole32 ReleaseStgMedium
make_lib user32 InsertMenuW
clang --target=x86_64-pc-windows-msvc -std=c11 -O2 -Wall -Wextra -Werror -ffreestanding -fno-builtin -fno-stack-protector -c shell.c -o "$tmp/shell.obj"
lld-link /dll /noentry /machine:x64 /nodefaultlib /dynamicbase /nxcompat /out:../HoI4NavigatorShell_v21.dll "$tmp/shell.obj" "$tmp/"*.lib /implib:"$tmp/output.lib"
