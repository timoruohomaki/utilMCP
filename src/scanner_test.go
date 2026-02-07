package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFolder_TextFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a text file.
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := ScanFolder(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Name != "hello.txt" {
		t.Errorf("expected name hello.txt, got %s", files[0].Name)
	}
	if files[0].Size != 11 {
		t.Errorf("expected size 11, got %d", files[0].Size)
	}
}

func TestScanFolder_SkipsBinaryFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a PNG-like binary file (PNG magic bytes).
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := os.WriteFile(filepath.Join(dir, "image.png"), png, 0644); err != nil {
		t.Fatal(err)
	}

	files, err := ScanFolder(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files (binary filtered), got %d", len(files))
	}
}

func TestScanFolder_SkipsOversizedFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a file larger than maxFileSize.
	bigFile := make([]byte, maxFileSize+1)
	for i := range bigFile {
		bigFile[i] = 'a'
	}
	if err := os.WriteFile(filepath.Join(dir, "big.txt"), bigFile, 0644); err != nil {
		t.Fatal(err)
	}

	files, err := ScanFolder(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files (oversized filtered), got %d", len(files))
	}
}

func TestScanFolder_SkipsDirectories(t *testing.T) {
	dir := t.TempDir()

	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := ScanFolder(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file (subdir skipped), got %d", len(files))
	}
}

func TestIsBinary(t *testing.T) {
	tests := []struct {
		mime   string
		binary bool
	}{
		{"text/plain", false},
		{"text/html", false},
		{"application/json", false},
		{"application/xml", false},
		{"application/javascript", false},
		{"application/octet-stream", true},
		{"image/png", true},
		{"application/pdf", true},
	}

	for _, tt := range tests {
		if got := isBinary(tt.mime); got != tt.binary {
			t.Errorf("isBinary(%q) = %v, want %v", tt.mime, got, tt.binary)
		}
	}
}
