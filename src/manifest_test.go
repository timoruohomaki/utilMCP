package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifestJSON(t *testing.T) {
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	m := Manifest{
		Files: []FileEntry{
			{Name: "test.txt", MimeType: "text/plain", Size: 42, LastModified: ts},
		},
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if len(decoded.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(decoded.Files))
	}
	if decoded.Files[0].Name != "test.txt" {
		t.Errorf("expected name test.txt, got %s", decoded.Files[0].Name)
	}
	if decoded.Files[0].Size != 42 {
		t.Errorf("expected size 42, got %d", decoded.Files[0].Size)
	}
}

func TestWriteManifest(t *testing.T) {
	dir := t.TempDir()
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	m := Manifest{
		Files: []FileEntry{
			{Name: "hello.txt", MimeType: "text/plain", Size: 11, LastModified: ts},
		},
	}

	if err := WriteManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	var loaded Manifest
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}

	if len(loaded.Files) != 1 || loaded.Files[0].Name != "hello.txt" {
		t.Errorf("unexpected manifest content: %+v", loaded)
	}
}
