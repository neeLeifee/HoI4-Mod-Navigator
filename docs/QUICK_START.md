# Quick start / Быстрый старт

Windows 10/11 x64 • unsigned preview • Russian settings UI.

1. Extract the ZIP and run `HoI4ModNavigator.exe` as your normal user. Verify the source and SHA-256 before running an unsigned binary.
2. Set the original game root (must contain `hoi4.exe`). Enter a name and existing root folder for each mod, then click «Добавить мод» (Add mod).
3. Click «Применить настройки» (Apply settings). Keep «Совместимый режим» (Compatibility mode) unchecked for dynamic menus.
4. Right-click a game folder or its background to open the corresponding mod folder. Right-click a file for copy, compare or reverse navigation. On Windows 11 choose Show more options. If needed, restart Explorer through Task Manager after finishing any active file operations.
5. To remove integration, click «Удалить интеграцию» (Remove integration) or run `remove_integration.cmd`. To revert to the static menu, enable Compatibility mode and Apply. Restart Explorer if needed. Do not overwrite a DLL that Explorer has loaded.

Settings and installed components: `%LOCALAPPDATA%\HoI4ModNavigator\`. Game files are not written by the navigation/copy commands. Missing destination folders can be created only in the configured mod; replacing a mod file requires confirmation. Always keep backups of mods.

---

1. Распакуйте архив, проверьте происхождение и SHA-256, запустите EXE от обычного пользователя.
2. Укажите папку игры с `hoi4.exe`, введите название и существующий корень мода, нажмите «Добавить мод».
3. Нажмите «Применить настройки». Для динамического меню не включайте совместимый режим.
4. ПКМ на фоне папки, папке или одиночном файле. В Windows 11 — «Показать дополнительные параметры». При необходимости перезапустите Проводник после завершения операций с файлами.
5. Для отката включите совместимый режим. Для удаления используйте кнопку «Удалить интеграцию» или `remove_integration.cmd`. Настройки и файлы игры/модов при удалении интеграции сохраняются.

Если меню не появилось или Explorer нестабилен — удалите интеграцию, перезапустите Проводник и используйте совместимый режим. Не отключайте SmartScreen глобально.
