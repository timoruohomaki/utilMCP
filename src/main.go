package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// All logging goes to stderr so stdout stays clean for MCP protocol.
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime)

	monitor := flag.Bool("monitor", false, "Log all MCP requests and responses to stderr")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: utilMCP <folder> [--monitor]\n\n")
		fmt.Fprintf(os.Stderr, "A read-only MCP server that exposes files in <folder> to AI applications.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	folder := flag.Arg(0)

	// Validate the folder exists and is a directory.
	info, err := os.Stat(folder)
	if err != nil {
		log.Fatalf("Cannot access folder: %v", err)
	}
	if !info.IsDir() {
		log.Fatalf("Not a directory: %s", folder)
	}

	// Scan files and build manifest.
	files, err := ScanFolder(folder)
	if err != nil {
		log.Fatalf("Failed to scan folder: %v", err)
	}

	manifest := Manifest{Files: files}
	if err := WriteManifest(folder, manifest); err != nil {
		log.Fatalf("Failed to write manifest: %v", err)
	}
	log.Printf("Manifest written with %d file(s)", len(files))

	// Create and start the MCP server.
	s := NewUtilMCPServer(folder, files, *monitor)

	log.Printf("utilMCP server starting (folder=%s, monitor=%v)", folder, *monitor)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
