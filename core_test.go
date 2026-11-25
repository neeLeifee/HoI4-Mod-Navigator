package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMapping(t *testing.T) {
	base := t.TempDir()
	game := filepath.Join(base, "Hearts of Iron IV")
	mod := filepath.Join(base, "Мой мод")
	cases := []struct {
		name, path, want string
		bad              bool
	}{
		{"root", game, mod, false},
		{"deep", filepath.Join(game, "gfx", "interface", "technologies", "1914"), filepath.Join(mod, "gfx", "interface", "technologies", "1914"), false},
		{"file", filepath.Join(game, "common", "units", "пехота.txt"), filepath.Join(mod, "common", "units", "пехота.txt"), false},
		{"prefix sibling", game + " backup", "", true},
		{"outside", filepath.Join(base, "other", "file.txt"), "", true},
		{"relative", "gfx", "", true},
		{"traversal", game + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "other", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, e := mapped(game, mod, c.path)
			if c.bad {
				if e == nil {
					t.Fatalf("accepted %q", got)
				}
				return
			}
			if e != nil || got != c.want {
				t.Fatalf("got %q %v, want %q", got, e, c.want)
			}
		})
	}
}
func TestConfig(t *testing.T) {
	base := t.TempDir()
	game := filepath.Join(base, "game")
	mod := filepath.Join(base, "mod")
	c := defaults()
	c.Game = game
	c.Mods = []Mod{{newID(), "WW1", mod}}
	mock := func(string) error { return nil }
	if e := validateConfig(c, mock); e != nil {
		t.Fatal(e)
	}
	cases := []struct {
		name string
		mods []Mod
	}{
		{"empty", nil}, {"same root", []Mod{{newID(), "WW1", game}}}, {"inside game", []Mod{{newID(), "WW1", filepath.Join(game, "mod")}}}, {"parent of game", []Mod{{newID(), "WW1", base}}}, {"duplicate name", []Mod{{newID(), "WW1", mod}, {newID(), "ww1", filepath.Join(base, "other")}}}, {"nested mods", []Mod{{newID(), "WW1", mod}, {newID(), "Other", filepath.Join(mod, "other")}}}, {"invalid id", []Mod{{"evil", "WW1", mod}}}, {"duplicate id", []Mod{{c.Mods[0].ID, "WW1", mod}, {c.Mods[0].ID, "Other", filepath.Join(base, "other")}}},
	}
	for _, x := range cases {
		t.Run(x.name, func(t *testing.T) {
			d := c
			d.Mods = x.mods
			if validateConfig(d, mock) == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
}

func TestReverseFileMapping(t *testing.T) {
	base := t.TempDir()
	game, mod := filepath.Join(base, "game"), filepath.Join(base, "mod")
	rel := filepath.Join("common", "units", "пехота.txt")
	dst, err := mapped(mod, game, filepath.Join(mod, rel))
	if err != nil || dst != filepath.Join(game, rel) {
		t.Fatalf("%q %v", dst, err)
	}
	if _, err := mapped(mod, game, filepath.Join(base, "other", rel)); err == nil {
		t.Fatal("outside mod accepted")
	}
}
func TestExplorerCommand(t *testing.T) {
	exe := `C:\Windows\explorer.exe`
	for _, c := range []struct {
		p          string
		selectFile bool
		want       string
	}{
		{`D:\Steam Library\Hearts of Iron IV`, false, `"C:\Windows\explorer.exe" /n,"D:\Steam Library\Hearts of Iron IV"`},
		{`D:\Моды, тест\common\file.txt`, true, `"C:\Windows\explorer.exe" /n,/select,"D:\Моды, тест\common\file.txt"`},
	} {
		if got := explorerCommand(exe, c.p, c.selectFile); got != c.want {
			t.Fatalf("%q != %q", got, c.want)
		}
	}
}
func TestReverseMenu(t *testing.T) {
	c := defaults()
	c.Game = "game"
	c.Mods = []Mod{{newID(), "WW1", "mod"}, {newID(), "WW2", "mod2"}}
	for _, m := range c.Mods {
		n := 0
		for _, item := range items(c, 2) {
			if item.action == "reverse-file" && item.id == m.ID {
				n++
				if len(item.roots) != 1 || item.roots[0] != m.Path {
					t.Fatal("wrong scope")
				}
			}
		}
		if n != 1 {
			t.Fatal("reverse file command missing or duplicated")
		}
	}
	c.Reverse = false
	for scope := 0; scope < 3; scope++ {
		for _, item := range items(c, scope) {
			if item.action == "reverse" || item.action == "reverse-file" {
				t.Fatal("disabled reverse visible")
			}
		}
	}
}

func TestAutoChoices(t *testing.T) {
	base := t.TempDir()
	c := defaults()
	c.Game = filepath.Join(base, "game")
	c.Mods = []Mod{{newID(), "FFU2", filepath.Join(base, "FFU2")}, {newID(), "other", filepath.Join(base, "other")}}
	cases := []struct {
		path   string
		action string
		count  int
		id     string
	}{
		{c.Game, "open", 2, c.Mods[0].ID},
		{filepath.Join(c.Game, "common", "units"), "open", 2, c.Mods[0].ID},
		{c.Mods[0].Path, "reverse", 1, c.Mods[0].ID},
		{filepath.Join(c.Mods[1].Path, "common", "units"), "reverse", 1, c.Mods[1].ID},
		{c.Mods[0].Path + "-backup", "", 0, ""},
		{base, "", 0, ""},
	}
	for _, v := range cases {
		got := autoChoices(c, v.path)
		if len(got) != v.count {
			t.Fatalf("%s: got %v", v.path, got)
		}
		if len(got) > 0 && (got[0].action != v.action || got[0].id != v.id) {
			t.Fatal("incorrect direction", got)
		}
	}
	c.Open = false
	if len(autoChoices(c, c.Game)) != 0 {
		t.Fatal("disabled open")
	}
	c.Reverse = false
	if len(autoChoices(c, c.Mods[0].Path)) != 0 || len(items(c, 0)) != 0 {
		t.Fatal("disabled reverse")
	}
}
