package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalModDiscovery(t *testing.T) {
	b := t.TempDir()
	c := defaults()
	c.Game = filepath.Join(b, "vanilla")
	c.LocalModsRoot = filepath.Join(b, "Documents", "Paradox Interactive", "Hearts of Iron IV", "mod")
	known := filepath.Join(c.LocalModsRoot, "FFU2")
	c.Mods = []Mod{{newID(), "FFU2", known}}
	root := filepath.Join(c.LocalModsRoot, "my local gdu")
	for _, rel := range []string{".", filepath.Join("gfx", "interface", "topbar"), filepath.Join("gfx", "interface", "topbar", "background.dds")} {
		src := filepath.Join(root, rel)
		got, err := localModRoot(c, src)
		if err != nil || got != root {
			t.Fatalf("root: %s %v", got, err)
		}
		dst, err := mapped(got, c.Game, src)
		if err != nil || dst != filepath.Join(c.Game, rel) {
			t.Fatalf("mapped: %s %v", dst, err)
		}
	}
	for _, p := range []string{c.LocalModsRoot, c.LocalModsRoot + "-backup", filepath.Join(b, "unrelated", "gfx"), filepath.Join(c.LocalModsRoot, "..", "gfx"), c.Game, filepath.Join(known, "gfx")} {
		if _, err := localModRoot(c, p); err == nil {
			t.Fatalf("accepted outside/explicit root: %q", p)
		}
	}
	p := filepath.Join(root, "gfx", "interface", "topbar")
	choice := autoChoices(c, p)
	if len(choice) != 1 || choice[0].action != "reverse-local" || choice[0].id != "local" {
		t.Fatal(choice)
	}
	choice = autoChoices(c, filepath.Join(known, "gfx"))
	if len(choice) != 1 || choice[0].id != c.Mods[0].ID {
		t.Fatal(choice)
	}
	choice = autoChoices(c, c.Game)
	if len(choice) != 1 || choice[0].id != c.Mods[0].ID {
		t.Fatal("auto mod became a target", choice)
	}
	c.Reverse = false
	if len(autoChoices(c, p)) != 0 {
		t.Fatal("disabled reverse available")
	}
	c.Mods = nil
	c.Reverse = true
	if err := validateConfig(c, func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(c)
	if err != nil || strings.Contains(string(encoded), "LocalModsRoot") {
		t.Fatal("derived path persisted", err)
	}
}

func TestLocalSnapshot(t *testing.T) {
	c := defaults()
	c.Game = `C:\Game`
	c.LocalModsRoot = `C:\Users\User\Documents\Paradox Interactive\Hearts of Iron IV\mod`
	c.Mods = []Mod{{newID(), "FFU2", c.LocalModsRoot + `\FFU2`}}
	for scope := 0; scope < 3; scope++ {
		found := map[string]int{}
		blocked := map[string]bool{}
		for _, m := range dynamicItems(c, scope) {
			if m.id != "local" {
				continue
			}
			found[m.action]++
			if m.action == "skip-local" {
				blocked[m.roots[0]] = true
			} else if m.roots[0] != c.LocalModsRoot {
				t.Fatal(m)
			}
		}
		if !blocked[c.Game] || !blocked[c.Mods[0].Path] {
			t.Fatal("explicit root precedence", blocked)
		}
		if scope == 2 {
			if found["reverse-local-file"] != 1 || found["compare-local"] != 1 || found["reverse-local"] != 0 {
				t.Fatal(found)
			}
		} else if found["reverse-local"] != 1 || found["compare-local"] != 0 {
			t.Fatal(found)
		}
		if found["open"] != 0 || found["copy"] != 0 {
			t.Fatal("unexpected write action", found)
		}
	}
	c.Reverse = false
	c.Compare = false
	for scope := 0; scope < 3; scope++ {
		for _, m := range dynamicItems(c, scope) {
			if m.id == "local" {
				t.Fatal("disabled command", m)
			}
		}
	}
}
