package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// FileEntry represents a single file's metadata in the manifest.
type FileEntry struct {
	Name         string    `json:"name"`
	MimeType     string    `json:"mimeType"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
}

// Manifest holds the list of files exposed by the server.
type Manifest struct {
	Files []FileEntry `json:"files"`
}

// WriteManifest serializes the manifest to JSON and writes it to the given directory.
func WriteManifest(dir string, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644)
}
