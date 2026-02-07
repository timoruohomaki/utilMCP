# Project Conventions

## Git Workflow

- After every code change, propose a commit with a draft message
- If the user accepts, stage the relevant files, commit, and push to origin
- Use concise commit messages that focus on the "why" rather than the "what"

## Documentation

- Keep README.md up to date when implementing new features or changing behavior

## Design Decisions

- **SDK**: mcp-go (github.com/mark3labs/mcp-go)
- **Transport**: stdio only, no HTTP/TLS
- **Directory scan**: flat (no recursion), binary and oversized files filtered out
- **Manifest**: manifest.json generated on startup with name, mimeType, size, lastModified
- **Resource URIs**: file:/// scheme with absolute paths
- **CLI**: `utilMCP /path/to/folder [--debug]`
- **Logging**: all output to stderr only; stdout reserved for MCP protocol
