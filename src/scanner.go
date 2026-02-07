package main

import (
	"net/http"
	"os"
	"strings"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

// ScanFolder reads the top-level entries of dir and returns metadata for
// eligible files (text-based, under the size limit). Binary and oversized
// files are silently skipped.
func ScanFolder(dir string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []FileEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		if info.Size() > maxFileSize {
			continue
		}

		fullPath := dir + "/" + e.Name()
		mime, err := detectMIME(fullPath)
		if err != nil {
			continue
		}

		if isBinary(mime) {
			continue
		}

		files = append(files, FileEntry{
			Name:         e.Name(),
			MimeType:     mime,
			Size:         info.Size(),
			LastModified: info.ModTime().UTC(),
		})
	}

	return files, nil
}

// detectMIME reads the first 512 bytes of a file and uses
// http.DetectContentType to determine the MIME type.
func detectMIME(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}

// isBinary returns true if the MIME type indicates a non-text file.
func isBinary(mime string) bool {
	return !strings.HasPrefix(mime, "text/") &&
		mime != "application/json" &&
		mime != "application/xml" &&
		mime != "application/javascript" &&
		mime != "application/x-yaml"
}
