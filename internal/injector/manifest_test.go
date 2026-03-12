package injector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestManifest_Has(t *testing.T) {
	m := &Manifest{
		Entries: []ManifestEntry{
			{TargetPath: "a/b", SourcePath: "/src/b", Mode: "symlink"},
			{TargetPath: "c/d", SourcePath: "/src/d", Mode: "copy"},
		},
	}

	if !m.Has("a/b") {
		t.Error("expected Has(a/b) = true")
	}
	if m.Has("x/y") {
		t.Error("expected Has(x/y) = false")
	}
}

func TestManifest_EntriesUnder(t *testing.T) {
	m := &Manifest{
		Entries: []ManifestEntry{
			{TargetPath: ".claude/skills/a.md"},
			{TargetPath: ".claude/skills/b.md"},
			{TargetPath: ".claude/skills/sub/c.md"},
			{TargetPath: ".claude/other.md"},
		},
	}

	got := m.EntriesUnder(".claude/skills")
	if len(got) != 2 {
		t.Fatalf("expected 2 direct children, got %d", len(got))
	}
	if got[0].TargetPath != ".claude/skills/a.md" || got[1].TargetPath != ".claude/skills/b.md" {
		t.Errorf("unexpected entries: %+v", got)
	}

	// No matches
	if len(m.EntriesUnder("nonexistent")) != 0 {
		t.Error("expected no entries under nonexistent prefix")
	}
}

func TestManifest_Add_New(t *testing.T) {
	m := &Manifest{}
	m.Add(ManifestEntry{TargetPath: "a", SourcePath: "/s/a", Mode: "symlink"})
	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m.Entries))
	}
	if m.Entries[0].TargetPath != "a" {
		t.Errorf("unexpected entry: %+v", m.Entries[0])
	}
}

func TestManifest_Add_Update(t *testing.T) {
	m := &Manifest{
		Entries: []ManifestEntry{{TargetPath: "a", SourcePath: "/old", Mode: "symlink"}},
	}
	m.Add(ManifestEntry{TargetPath: "a", SourcePath: "/new", Mode: "copy"})
	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry after update, got %d", len(m.Entries))
	}
	if m.Entries[0].SourcePath != "/new" || m.Entries[0].Mode != "copy" {
		t.Errorf("entry not updated: %+v", m.Entries[0])
	}
}

func TestManifest_Remove(t *testing.T) {
	m := &Manifest{
		Entries: []ManifestEntry{
			{TargetPath: "a"},
			{TargetPath: "b"},
			{TargetPath: "c"},
		},
	}
	m.Remove("b")
	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}
	if m.Has("b") {
		t.Error("b should have been removed")
	}

	// Remove nonexistent is a no-op
	m.Remove("z")
	if len(m.Entries) != 2 {
		t.Error("removing nonexistent should not change entries")
	}
}

func TestManifest_TargetPaths(t *testing.T) {
	m := &Manifest{
		Entries: []ManifestEntry{
			{TargetPath: "a"},
			{TargetPath: "b"},
			{TargetPath: "a"}, // duplicate
			{TargetPath: "c"},
		},
	}
	paths := m.TargetPaths()
	if len(paths) != 3 {
		t.Fatalf("expected 3 distinct paths, got %d: %v", len(paths), paths)
	}
	want := []string{"a", "b", "c"}
	for i, w := range want {
		if paths[i] != w {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], w)
		}
	}
}

func TestLoadManifest_MissingFile(t *testing.T) {
	m := LoadManifest(t.TempDir())
	if len(m.Entries) != 0 {
		t.Errorf("expected empty manifest for missing file, got %d entries", len(m.Entries))
	}
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	mDir := filepath.Join(dir, manifestDir)
	if err := os.MkdirAll(mDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mDir, manifestFile), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	m := LoadManifest(dir)
	if len(m.Entries) != 0 {
		t.Errorf("expected empty manifest for invalid JSON, got %d entries", len(m.Entries))
	}
}

func TestLoadManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	mDir := filepath.Join(dir, manifestDir)
	if err := os.MkdirAll(mDir, 0755); err != nil {
		t.Fatal(err)
	}
	want := &Manifest{
		Entries: []ManifestEntry{{TargetPath: "x", SourcePath: "/y", Mode: "copy"}},
	}
	data, _ := json.Marshal(want)
	if err := os.WriteFile(filepath.Join(mDir, manifestFile), data, 0644); err != nil {
		t.Fatal(err)
	}
	got := LoadManifest(dir)
	if len(got.Entries) != 1 || got.Entries[0].TargetPath != "x" {
		t.Errorf("unexpected manifest: %+v", got)
	}
}

func TestSaveManifest(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{
		Entries: []ManifestEntry{{TargetPath: "a", SourcePath: "/b", Mode: "symlink"}},
	}
	if err := SaveManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(filepath.Join(dir, manifestDir, manifestFile))
	if err != nil {
		t.Fatal(err)
	}
	var loaded Manifest
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 1 || loaded.Entries[0].TargetPath != "a" {
		t.Errorf("saved manifest mismatch: %+v", loaded)
	}
}

func TestClearManifest(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{Entries: []ManifestEntry{{TargetPath: "a"}}}
	if err := SaveManifest(dir, m); err != nil {
		t.Fatal(err)
	}
	if err := ClearManifest(dir); err != nil {
		t.Fatal(err)
	}
	// File should be gone
	if _, err := os.Stat(filepath.Join(dir, manifestDir, manifestFile)); !os.IsNotExist(err) {
		t.Error("manifest file should have been removed")
	}

	// Clearing again should not error
	if err := ClearManifest(dir); err != nil {
		t.Errorf("second clear should not error: %v", err)
	}
}
