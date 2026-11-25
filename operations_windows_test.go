package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsPaths(t *testing.T) {
	g := `D:\Steam\Hearts of Iron IV`
	m := `C:\Mods\WW1`
	p := `d:\STEAM\Hearts of Iron IV\gfx\interface`
	v, e := mapped(g, m, p)
	if e != nil || !strings.EqualFold(v, `C:\Mods\WW1\gfx\interface`) {
		t.Fatalf("%q %v", v, e)
	}
	for _, p := range []string{`D:\Steam\Hearts of Iron IV backup\x`, `C:\Other\x`, `D:\Steam\Hearts of Iron IV\..\x`} {
		if _, e := mapped(g, m, p); e == nil {
			t.Fatal(p)
		}
	}
	if localPath(`C:\Mods\x.txt:stream`) == nil {
		t.Fatal("ADS accepted")
	}
}
func TestEnsureAndReplace(t *testing.T) {
	r := t.TempDir()
	d := filepath.Join(r, "deep", "folder")
	if e := ensureDir(r, d, false); e == nil {
		t.Fatal("created without permission")
	}
	if e := ensureDir(r, d, true); e != nil {
		t.Fatal(e)
	}
	a, b := filepath.Join(d, "a"), filepath.Join(d, "b")
	os.WriteFile(a, []byte("a"), 0600)
	os.WriteFile(b, []byte("b"), 0600)
	if e := replace(a, b, false); e == nil {
		t.Fatal("overwrote without permission")
	}
	if v, _ := os.ReadFile(b); string(v) != "b" {
		t.Fatal("changed existing file")
	}
	if e := replace(a, b, true); e != nil {
		t.Fatal(e)
	}
	if v, _ := os.ReadFile(b); string(v) != "a" {
		t.Fatal("wrong contents")
	}
}
func TestMenuGeneration(t *testing.T) {
	c := defaults()
	c.Game = `C:\Game`
	c.Mods = []Mod{{newID(), "WW1", `C:\Mods\WW1`}}
	if len(items(c, 0)) != 1 || len(items(c, 1)) != 2 || len(items(c, 2)) != 4 {
		t.Fatal("bad commands")
	}
	c.Reverse = false
	if len(items(c, 0)) != 1 {
		t.Fatal("bad single command")
	}
	c.Open = false
	c.Copy = false
	c.Compare = false
	if len(items(c, 2)) != 0 {
		t.Fatal("disabled commands still present")
	}
}
