# utilMCP

A minimal, read-only MCP server written in Go that exposes files within a folder to AI applications such as Claude Desktop.

## What is MCP?

[MCP (Model Context Protocol)](https://modelcontextprotocol.io) is an open standard for connecting AI applications to external data sources and tools. It provides a standardized way for AI hosts (like Claude Desktop, VS Code, or custom agents) to communicate with servers that supply context -- files, database records, API responses, executable functions, and more.

### Architecture

MCP follows a client-server model with three participants:

- **Host**: The AI application (e.g. Claude Desktop) that coordinates connections
- **Client**: A component inside the host that maintains a 1:1 connection to a server
- **Server**: A program that provides context to clients via a defined protocol

```
┌─────────────────────────────────┐
│          MCP Host               │
│       (Claude Desktop)          │
│                                 │
│  ┌──────────┐   ┌──────────┐   │
│  │ Client 1 │   │ Client 2 │   │
│  └────┬─────┘   └────┬─────┘   │
└───────┼──────────────┼──────────┘
        │              │
   ┌────▼─────┐   ┌────▼─────┐
   │ Server A │   │ Server B │
   │(utilMCP) │   │ (other)  │
   └──────────┘   └──────────┘
```

### Protocol Basics

MCP uses [JSON-RPC 2.0](https://www.jsonrpc.org/) for message encoding. Communication begins with an initialization handshake where client and server negotiate capabilities, then proceeds with request/response exchanges.

### Transport

MCP supports two transports:

| Transport | Description | Use Case |
|---|---|---|
| **stdio** | Server is launched as a subprocess; messages are exchanged over stdin/stdout, delimited by newlines | Local servers (Claude Desktop) |
| **Streamable HTTP** | Server runs independently; messages are sent via HTTP POST with optional SSE streaming | Remote servers |

Claude Desktop launches local MCP servers using **stdio**: it spawns the server process and communicates via stdin/stdout. The server must never write non-protocol output to stdout (use stderr for logging).

### Server Primitives

An MCP server can expose three types of capabilities:

| Primitive | Purpose | Discovery | Use |
|---|---|---|---|
| **Resources** | Read-only data (file contents, DB records, API responses) | `resources/list` | `resources/read` |
| **Tools** | Executable functions the LLM can invoke (with user approval) | `tools/list` | `tools/call` |
| **Prompts** | Reusable interaction templates | `prompts/list` | `prompts/get` |

## Design

### Dependencies

- [mcp-go](https://github.com/mark3labs/mcp-go) -- Go SDK for the Model Context Protocol

### Scope

- **Transport**: stdio only (local desktop use, no TLS/auth needed)
- **Directory**: flat scan of the target folder (no recursion into subdirectories)
- **Read-only**: no write operations; the server only exposes file contents
- **File filtering**: binary files and files over 10 MB are skipped

### manifest.json

On startup, utilMCP scans the target folder and writes a `manifest.json` containing metadata for each eligible file:

```json
{
  "files": [
    {
      "name": "example.txt",
      "mimeType": "text/plain",
      "size": 1024,
      "lastModified": "2025-01-15T10:30:00Z"
    }
  ]
}
```

This manifest is regenerated each time the server starts.

### Resource URIs

Files are exposed using `file:///` URIs pointing to their absolute paths on disk.

### CLI Usage

```
utilMCP /path/to/folder [--debug]
```

| Argument | Description |
|---|---|
| `/path/to/folder` | Required. The folder whose files to expose |
| `--debug` | Optional. Log all MCP requests and responses to stderr for debugging |

## Minimal Features for Claude Desktop Compatibility

To function as a valid MCP server that Claude Desktop can connect to, utilMCP must implement:

### Required Protocol Methods

1. **`initialize`** -- Respond to the client's initialization handshake, negotiate protocol version, and declare server capabilities
2. **`notifications/initialized`** -- Accept the client's ready notification

### Required Capabilities (for this project)

Since utilMCP is a read-only file server, it implements:

- **`resources/list`** -- Return a list of available files in the configured folder
- **`resources/read`** -- Return the contents of a requested file
- **`tools/list`** -- Expose `list_files` and `read_file` tools
- **`tools/call`** -- Execute tool calls and return results

### Exposed Tools

| Tool | Description |
|---|---|
| `list_files` | Returns the manifest (all available files with metadata) |
| `read_file` | Reads and returns the contents of a file by name |

### What This Server Does NOT Need

- Prompts (no reusable templates needed)
- Sampling/Elicitation (no LLM requests or user input from server side)
- Streamable HTTP transport (stdio is sufficient for Claude Desktop)
- Write operations (read-only by design)

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)

## Building

### macOS

```bash
go build -o utilMCP ./src/
go test ./src/
```

### Windows 11

```powershell
go build -o utilMCP.exe ./src/
go test ./src/
```

## Installation

### macOS

1. Build the binary (see above).
2. Move it to a location on your PATH, or note its absolute path:

```bash
# Option A: install to a standard location
cp utilMCP /usr/local/bin/

# Option B: use the absolute path directly in the config
# e.g. /Users/yourname/projects/utilMCP/utilMCP
```

3. Edit the Claude Desktop configuration file:

```bash
# Create/open the config file
open ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

4. Add the server entry:

```json
{
  "mcpServers": {
    "utilMCP": {
      "command": "/usr/local/bin/utilMCP",
      "args": ["/Users/yourname/Documents/my-files"]
    }
  }
}
```

With debug logging:

```json
{
  "mcpServers": {
    "utilMCP": {
      "command": "/usr/local/bin/utilMCP",
      "args": ["/Users/yourname/Documents/my-files", "--debug"]
    }
  }
}
```

5. Restart Claude Desktop. Debug logs are written to `~/Library/Logs/Claude/mcp-server-utilMCP.log`.

### Windows 11

1. Build the binary (see above).
2. Place `utilMCP.exe` in a known location, e.g. `C:\Tools\utilMCP.exe`.
3. Edit the Claude Desktop configuration file:

```powershell
# Create/open the config file
notepad "$env:APPDATA\Claude\claude_desktop_config.json"
```

4. Add the server entry (use forward slashes or escaped backslashes in JSON):

```json
{
  "mcpServers": {
    "utilMCP": {
      "command": "C:/Tools/utilMCP.exe",
      "args": ["C:/Users/yourname/Documents/my-files"]
    }
  }
}
```

With debug logging:

```json
{
  "mcpServers": {
    "utilMCP": {
      "command": "C:/Tools/utilMCP.exe",
      "args": ["C:/Users/yourname/Documents/my-files", "--debug"]
    }
  }
}
```

5. Restart Claude Desktop. Debug logs are written to `%APPDATA%\Claude\Logs\mcp-server-utilMCP.log`.

### File Access Restrictions

utilMCP uses `file:///` URIs to reference files on disk. Be aware of the following restrictions:

- **The server can only read files that the running user has OS-level read permission for.** It runs as your user account, so standard file permissions apply.
- **macOS**: if the exposed folder is in a protected location (Desktop, Documents, Downloads, iCloud Drive, or external volumes), macOS may prompt for permission or block access entirely. The simplest approach is to use a folder outside these protected areas, or grant Full Disk Access to the terminal/shell that launches the server (System Settings > Privacy & Security > Full Disk Access).
- **Windows**: folders protected by Windows Security (e.g. `C:\Windows`, `C:\Program Files`) or OneDrive-managed paths may require elevated permissions or specific access grants.
- **Paths with spaces** work correctly -- they are handled by the Go standard library and passed through as-is in the JSON config.
- **Symlinks**: the server follows symlinks when reading files, but the resource URI will reflect the original (symlinked) path. Ensure the symlink target is also readable.
- **Network paths** (UNC paths on Windows, NFS/SMB mounts on macOS) are not explicitly supported and may behave unpredictably.

## License

See [LICENSE](LICENSE) for details.
