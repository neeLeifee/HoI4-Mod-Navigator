# Changelog

## 2.1.0-beta.1 — automatic local mod detection

- Detect local HoI4 mods under the Windows Documents known folder without registering every mod.
- Read-only reverse folder/file navigation and VS Code comparison; write destinations remain explicitly configured.
- Keep explicit roots authoritative and reject unrelated paths.
- Preserve the PowerShell quoting fix in CI.
- Version the native DLL separately; restart Explorer after upgrading.

## 2.0.0-beta.1 — repository preview

- Native x64 COM context-menu handler: contextual labels and no commands outside configured roots in dynamic mode.
- Up to 20 configured mods; vanilla-to-mod and mod-to-vanilla navigation, file selection, copying and VS Code comparison.
- Legacy menu compatibility mode and emergency registry cleanup.
- Repository packaging, bilingual documentation, illustrated demos and source-based CI builds.
- Corrected a stale assertion in the Windows-only test for the legacy background menu (one dispatcher, not two); no navigation changes.

## 1.5.1

- Removed experimental background filtering that could hide every command.
- Restored the automatic vanilla/mod dispatcher.

## 1.4–1.5

- Refreshed Win32 settings UI and native Common Item Dialog folder picker.
- Experimented with context-specific labels on the folder background.

## 1.0–1.3

- Initial navigation, folder creation, copy confirmation, reverse file selection, VS Code comparison and multiple mod support.
