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

	debug := flag.Bool("debug", false, "Log all MCP requests and responses to stderr")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: utilMCP <folder> [--debug]\n\n")
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
	s := NewUtilMCPServer(folder, files, *debug)

	fmt.Fprintln(os.Stderr, `utilMCP - A read-only MCP server for AI applications.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`)

	log.Printf("utilMCP server starting (folder=%s, debug=%v)", folder, *debug)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
