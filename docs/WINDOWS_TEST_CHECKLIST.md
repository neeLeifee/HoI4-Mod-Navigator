# Manual Windows release gate

Record Windows edition/build, architecture, date, application tag/commit, and actual results. Automated CI does not replace these checks. Start with disposable roots and a placeholder `hoi4.exe` (never execute that placeholder), not valuable game/mod data.

- [ ] Clean install as normal user, then migrate real v1.5.1/v2.0 settings if available.
- [ ] Modern folder picker, add/edit/remove mods, Unicode and spaces, settings persistence.
- [ ] Background and selected-folder menu inside vanilla: a named destination for each of 1, 2 and 20 configured mods.
- [ ] Inside each mod: only reverse destination for that mod. Unrelated directory and similarly named sibling: no dynamic commands.
- [ ] On vanilla file: open parent in mod, copy, compare. On mod file: select original, compare. Multiple selection shows no native commands.
- [ ] Navigate into missing nested mod folders with creation enabled; disable creation and verify refusal. Original folder remains open; new window requested.
- [ ] Reverse navigation when target is absent: understandable error, no vanilla writes.
- [ ] Copy into test mod, cancel overwrite, confirm overwrite, compare hashes. Compare in VS Code if installed; missing editor should produce an error.
- [ ] Reject nested roots, non-local paths, junctions/reparse points, and prefix siblings.
- [ ] Windows 11 classic menu via Show more options. Windows 10 classic menu.
- [ ] Reapply settings; restart Explorer; check responsiveness and no duplicate menu entries.
- [ ] Compatibility mode, remove integration, emergency CMD, restart Explorer, then reinstall.
- [ ] Verify unsigned warning and hashes; do not disable system protection.

Release only after documenting which OS configurations were actually tested. If an item fails, leave release as preview/draft and report it.


## v2.1 automatic local mod discovery

- Restart Explorer after applying the new version.
- Open an unregistered local mod under the actual Windows Documents/Paradox Interactive/Hearts of Iron IV/mod directory. Test folder background, selected folder, original file selection and VS Code comparison.
- Verify a configured local mod shows no duplicate automatic commands; verify vanilla shows only configured destinations.
- Create another local mod directory after applying settings and confirm that it works without reapplying.
- Verify the parent mod container, unrelated paths and standalone .mod descriptor files do not offer dynamic reverse commands. Disable Reverse/Compare and check the corresponding automatic actions disappear.
- Test compatibility mode background navigation and rollback.
