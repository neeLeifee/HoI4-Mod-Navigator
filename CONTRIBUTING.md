# Contributing

- Open an issue describing the change before a large refactor.
- Keep the native handler small: it chooses menu entries and launches the EXE; filesystem operations stay in Go.
- Run `bash scripts/build.sh` under Linux/WSL with Go, Clang and LLD installed; run `dist/windows-tests.exe -test.v` on Windows x64.
- For Explorer-related changes, follow [the manual smoke-test checklist](docs/WINDOWS_TEST_CHECKLIST.md). Do not test with valuable mod data before checking in disposable folders.
- Include tests, avoid committing DLL/EXE outputs, and keep docs and screenshots honest. Update the DLL filename when changing its binary to avoid replacing a loaded DLL.
- Pull requests must not include game assets, mod contents, personal paths, or credentials.
