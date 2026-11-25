#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
for cmd in go clang lld-link zip sha256sum; do command -v "$cmd" >/dev/null || { echo "Missing tool: $cmd" >&2; exit 1; }; done
mkdir -p dist
bash native/build.sh
clang --target=x86_64-pc-windows-msvc -x assembler -c resource.asm.txt -o icon_windows_amd64.syso
go test -count=1 ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o dist/windows-tests.exe
clang -std=c11 -fshort-wchar -Wall -Wextra -Werror -fsanitize=address,undefined native/test_shell.c -o dist/native-test
./dist/native-test
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -H windowsgui' -o dist/HoI4ModNavigator.exe .
cp remove_integration.cmd dist/
cp README.md dist/README.md
cp docs/QUICK_START.md dist/QUICK_START.md
cp LICENSE dist/LICENSE
(
 cd dist
 rm -f HoI4ModNavigator-Windows-x64.zip SHA256SUMS.txt
 # The ZIP contains only end-user files, not test executables.
 sha256sum HoI4ModNavigator.exe > SHA256SUMS.txt
 zip -q HoI4ModNavigator-Windows-x64.zip HoI4ModNavigator.exe remove_integration.cmd QUICK_START.md LICENSE SHA256SUMS.txt
 sha256sum HoI4ModNavigator-Windows-x64.zip >> SHA256SUMS.txt
)
echo 'Build complete: dist/HoI4ModNavigator-Windows-x64.zip'
