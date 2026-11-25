package main

import (
	"fmt"
	"strings"
)

const baseKey = `Software\Classes\`
const ownVerb = "HoI4ModNavigator.3C81A9"
const maxMenuItems = 80 // 20 mods * 4 file actions
var scopes = []string{`Directory\Background\shell\`, `Directory\shell\`, `*\shell\`}

type regEntry struct{ path, name, value string }

func quoted(s string) string { return `"` + s + `"` }
func condition(roots []string) string {
	var a []string
	for _, root := range roots {
		a = append(a, `System.ParsingPath:="`+root+`"`, `System.ParsingPath:~<"`+root+`\"`)
	}
	return "(" + strings.Join(a, " OR ") + ")"
}
func verbKey(scope, index int) string {
	return fmt.Sprintf(`%s%s%s.cmd%03d`, baseKey, scopes[scope], ownVerb, index)
}
func removalKeys() []string {
	var keys []string
	for i, s := range scopes {
		keys = append(keys, baseKey+s+ownVerb) // v1.0/v1.1 cascade
		for j := 0; j < maxMenuItems; j++ {
			keys = append(keys, verbKey(i, j))
		}
	}
	return keys
}
func menuEntries(c Config, exe string) ([]regEntry, error) {
	if strings.ContainsAny(exe, "%\"\x00\r\n") {
		return nil, fmt.Errorf("Недопустимый путь программы")
	}
	var entries []regEntry
	for i := range scopes {
		all := items(c, i)
		if len(all) > maxMenuItems {
			return nil, fmt.Errorf("Слишком много команд")
		}
		placeholder := "%1"
		if i == 0 {
			placeholder = "%V"
		}
		for j, m := range all {
			key := verbKey(i, j)
			// Each verb is independently executable; no parent or command store needed.
			entries = append(entries, regEntry{key, "", m.label}, regEntry{key, "MUIVerb", m.label}, regEntry{key, "Icon", quoted(exe) + ",0"}, regEntry{key + `\command`, "", quoted(exe) + " action " + m.action + " " + m.id + " " + quoted(placeholder)})
			if c.Filter && i != 0 {
				entries = append(entries, regEntry{key, "AppliesTo", condition(m.roots)})
			}
			if i == 2 {
				entries = append(entries, regEntry{key, "MultiSelectModel", "Single"})
			}
		}
	}
	return entries, nil
}
