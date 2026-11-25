package main

import (
	"fmt"
	"strings"
	"unicode/utf16"
)

const shellKey = `Software\HoI4ModNavigator\ShellV2`
const dllName = "HoI4NavigatorShell_v20.dll"

var classIDs = []string{"{76AE4D21-CAD1-4B8E-A79E-8555CAB57101}", "{76AE4D21-CAD1-4B8E-A79E-8555CAB57102}", "{76AE4D21-CAD1-4B8E-A79E-8555CAB57103}"}
var handlerScopes = []string{`Directory\Background`, `Directory`, `*`}

func handlerKey(i int) string {
	return baseKey + handlerScopes[i] + `\shellex\ContextMenuHandlers\` + classIDs[i]
}
func classKey(i int) string { return baseKey + `CLSID\` + classIDs[i] }
func dynamicItems(c Config, scope int) []menuItem {
	if scope == 0 {
		return items(c, 1)
	}
	return items(c, scope)
}
func snapshot(c Config, exe string) (string, error) {
	var b strings.Builder
	add := func(s string) error {
		if strings.ContainsAny(s, "\r\n\x00") {
			return fmt.Errorf("Недопустимые символы в настройках меню")
		}
		b.WriteString(s)
		b.WriteByte('\n')
		return nil
	}
	for _, s := range []string{"H4N2", exe} {
		if e := add(s); e != nil {
			return "", e
		}
	}
	for scope := 0; scope < 3; scope++ {
		for _, m := range dynamicItems(c, scope) {
			for _, root := range m.roots {
				for _, s := range []string{fmt.Sprint(scope), root, m.label, m.action, m.id} {
					if e := add(s); e != nil {
						return "", e
					}
				}
			}
		}
	}
	result := b.String()
	if (len(utf16.Encode([]rune(result)))+1)*2 > 1024*1024 {
		return "", fmt.Errorf("Настройки меню слишком большие")
	}
	return result, nil
}
func dynamicEntries(c Config, exe, dll string) ([]regEntry, error) {
	s, e := snapshot(c, exe)
	if e != nil {
		return nil, e
	}
	out := []regEntry{{shellKey, "Menu", s}}
	for i := range classIDs {
		k := classKey(i)
		out = append(out, regEntry{k, "", "HoI4 Mod Navigator"}, regEntry{k + `\InprocServer32`, "", dll}, regEntry{k + `\InprocServer32`, "ThreadingModel", "Apartment"})
	}
	// Register handlers last, once COM classes and menu snapshot are available.
	for i, id := range classIDs {
		out = append(out, regEntry{handlerKey(i), "", id})
	}
	return out, nil
}
