package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewUtilMCPServer creates an MCP server that exposes files in folder as
// read-only resources and tools.
func NewUtilMCPServer(folder string, files []FileEntry, monitor bool) *server.MCPServer {
	opts := []server.ServerOption{
		server.WithResourceCapabilities(false, false),
		server.WithToolCapabilities(false),
	}

	if monitor {
		hooks := &server.Hooks{}
		hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
			log.Printf("[monitor] tool/call: %s args=%v", message.Params.Name, message.Params.Arguments)
		})
		hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
			log.Printf("[monitor] tool/call result: isError=%v", result.IsError)
		})
		opts = append(opts, server.WithHooks(hooks))
	}

	s := server.NewMCPServer("utilMCP", "0.1.0", opts...)

	registerResources(s, folder, files)
	registerTools(s, folder, files)

	return s
}

func registerResources(s *server.MCPServer, folder string, files []FileEntry) {
	for _, f := range files {
		absPath, _ := filepath.Abs(filepath.Join(folder, f.Name))
		uri := "file:///" + absPath

		resource := mcp.NewResource(
			uri,
			f.Name,
			mcp.WithResourceDescription(fmt.Sprintf("File: %s (%s, %d bytes)", f.Name, f.MimeType, f.Size)),
			mcp.WithMIMEType(f.MimeType),
		)

		s.AddResource(resource, makeResourceHandler(absPath))
	}
}

func makeResourceHandler(path string) server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     string(data),
			},
		}, nil
	}
}

func registerTools(s *server.MCPServer, folder string, files []FileEntry) {
	// list_files tool
	listFilesTool := mcp.NewTool(
		"list_files",
		mcp.WithDescription("List all available files in the exposed folder"),
	)
	s.AddTool(listFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := json.MarshalIndent(files, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialize file list"), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	})

	// read_file tool
	readFileTool := mcp.NewTool(
		"read_file",
		mcp.WithDescription("Read the contents of a file by name"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the file to read")),
	)
	s.AddTool(readFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: name"), nil
		}

		// Validate the file is in our manifest
		found := false
		for _, f := range files {
			if f.Name == name {
				found = true
				break
			}
		}
		if !found {
			return mcp.NewToolResultError(fmt.Sprintf("file not found: %s", name)), nil
		}

		absPath, _ := filepath.Abs(filepath.Join(folder, name))
		data, err := os.ReadFile(absPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})
}
