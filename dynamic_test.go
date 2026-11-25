package main

import (
	"strings"
	"testing"
)

func TestDynamicMenuPlan(t *testing.T) {
	c := defaults()
	c.Game = `C:\Games\HOI4`
	c.Mods = []Mod{{"0123456789abcdef01234567", "FFU2", `D:\Mods\FFU2`}, {"abcdef0123456789abcdef01", "Second", `D:\Mods\Second`}}
	e, err := dynamicEntries(c, `C:\Nav\Navigator.exe`, `C:\Nav\Shell.dll`)
	if err != nil {
		t.Fatal(err)
	}
	if len(e) != 13 {
		t.Fatal(len(e))
	}
	if e[0].path != shellKey {
		t.Fatal(e[0])
	}
	s := strings.Split(e[0].value, "\n")
	if s[0] != "H4N2" || s[1] != `C:\Nav\Navigator.exe` {
		t.Fatal(s[:2])
	}
	type command struct{ scope, root, label, action, id string }
	var cmds []command
	for i := 2; i+4 < len(s)-1; i += 5 {
		cmds = append(cmds, command{s[i], s[i+1], s[i+2], s[i+3], s[i+4]})
	}
	if len(cmds) != 18 {
		t.Fatal(len(cmds), cmds)
	}
	for _, x := range cmds {
		if x.scope == "0" && x.root == c.Game && x.action != "open" {
			t.Fatal(x)
		}
		if x.scope == "0" && x.root != c.Game && x.action != "reverse" {
			t.Fatal(x)
		}
		if x.root == "" || x.action == "auto-folder" {
			t.Fatal(x)
		}
	}
	for i := 0; i < 3; i++ {
		if e[len(e)-3+i].path != handlerKey(i) || e[len(e)-3+i].value != classIDs[i] {
			t.Fatal(e)
		}
	}
	c.Open = false
	c.Reverse = false
	c.Copy = false
	c.Compare = false
	snap, err := snapshot(c, `C:\Nav.exe`)
	if err != nil || snap != "H4N2\nC:\\Nav.exe\n" {
		t.Fatal(snap, err)
	}
	c.Mods[0].Name = "bad\nname"
	c.Open = true
	if _, err = snapshot(c, `C:\Nav.exe`); err == nil {
		t.Fatal("newline accepted")
	}
}
