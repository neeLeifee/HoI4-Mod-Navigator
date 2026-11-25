package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFlatMenuRegistration(t *testing.T) {
	c := defaults()
	c.Filter = false
	c.Game = `D:\Steam\Hearts of Iron IV`
	c.Mods = []Mod{{newID(), "FFU2", `D:\Mods\FFU2`}}
	exe := `C:\Users\Test User\AppData\Local\HoI4ModNavigator\HoI4ModNavigator.exe`
	plan, e := menuEntries(c, exe)
	if e != nil {
		t.Fatal(e)
	}
	vals := map[string]string{}
	for _, v := range plan {
		vals[v.path+"|"+v.name] = v.value
		if strings.Contains(v.path, "ExtendedSubCommandsKey") {
			t.Fatal("cascade found")
		}
	}
	keys := map[string]bool{}
	for _, k := range removalKeys() {
		keys[k] = true
	}
	for i := range scopes {
		for j, m := range items(c, i) {
			k := verbKey(i, j)
			ph := "%1"
			if i == 0 {
				ph = "%V"
			}
			if !keys[k] {
				t.Fatal("missing cleanup")
			}
			if vals[k+"|MUIVerb"] != m.label || vals[k+"|"] != m.label {
				t.Fatal("wrong label", k)
			}
			want := quoted(exe) + " action " + m.action + " " + m.id + " " + quoted(ph)
			if vals[k+`\command|`] != want {
				t.Fatal("wrong command", k)
			}
			if _, ok := vals[k+"|AppliesTo"]; ok {
				t.Fatal("unexpected filter")
			}
		}
		if !keys[baseKey+scopes[i]+ownVerb] {
			t.Fatal("missing legacy cleanup")
		}
	}
	if len(plan) != 32 {
		t.Fatalf("unexpected entries: %d", len(plan))
	}
}
func TestMenuMaxAndFilter(t *testing.T) {
	c := defaults()
	c.Filter = true
	c.Game = `C:\Game`
	for i := 0; i < 20; i++ {
		c.Mods = append(c.Mods, Mod{newID(), "mod", `C:\Mod`})
	}
	p, e := menuEntries(c, `C:\Program.exe`)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, v := range p {
		if v.path == verbKey(2, 79)+`\command` {
			found = true
		}
		if v.name == "AppliesTo" && v.value == "" {
			t.Fatal("empty filter")
		}
	}
	if !found {
		t.Fatal("last command missing")
	}
	c.Mods = append(c.Mods, c.Mods[0])
	if _, e := menuEntries(c, `C:\Program.exe`); e == nil {
		t.Fatal("limit not enforced")
	}
}

func TestAdaptiveMenuRegistration(t *testing.T) {
	c := defaults()
	c.Game = `D:\Steam\Hearts of Iron IV`
	c.Mods = []Mod{{newID(), "FFU2", `D:\Mods\FFU2`}, {newID(), "Other", `D:\Mods\Other`}}
	plan, err := menuEntries(c, `C:\Navigator.exe`)
	if err != nil {
		t.Fatal(err)
	}
	vals := map[string]string{}
	for _, v := range plan {
		vals[v.path+"|"+v.name] = v.value
	}
	bg := verbKey(0, 0)
	if vals[bg+`\command|`] != `"C:\Navigator.exe" action auto-folder auto "%V"` {
		t.Fatal("background dispatch incorrect")
	}
	if _, ok := vals[bg+"|AppliesTo"]; ok {
		t.Fatal("background must not be filtered")
	}
	if len(items(c, 0)) != 1 {
		t.Fatal("background has redundant commands")
	}
	for scope := 1; scope <= 2; scope++ {
		for i, m := range items(c, scope) {
			want := condition(m.roots)
			if vals[verbKey(scope, i)+"|AppliesTo"] != want {
				t.Fatal("wrong filter", m.action, m.label)
			}
		}
	}
	want := `(System.ParsingPath:="D:\Mods\FFU2" OR System.ParsingPath:~<"D:\Mods\FFU2\")`
	if condition([]string{c.Mods[0].Path}) != want {
		t.Fatal("bad AQS root boundary")
	}
}

func TestLegacyNamedBackgroundIgnored(t *testing.T) {
	c := defaults()
	if err := json.Unmarshal([]byte(`{"NamedBackground":true}`), &c); err != nil {
		t.Fatal(err)
	}
	c.Game = `D:\Game`
	c.Mods = []Mod{{newID(), "FFU2", `D:\Mods\FFU2`}}
	for _, filter := range []bool{false, true} {
		c.Filter = filter
		bg := items(c, 0)
		if len(bg) != 1 || bg[0].action != "auto-folder" || !strings.Contains(bg[0].label, "FFU2") {
			t.Fatal(bg)
		}
		plan, err := menuEntries(c, `C:\Navigator.exe`)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, v := range plan {
			if strings.HasPrefix(v.path, baseKey+scopes[0]) {
				if v.name == "AppliesTo" {
					t.Fatal("background filtered")
				}
				if v.path == verbKey(0, 0)+`\command` && v.value == `"C:\Navigator.exe" action auto-folder auto "%V"` {
					found = true
				}
			}
		}
		if !found {
			t.Fatal("missing auto command")
		}
	}
}
