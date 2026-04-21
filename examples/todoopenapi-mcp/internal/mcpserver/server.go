package mcpserver

import (
	"github.com/mark3labs/mcp-go/server"
	"todoopenapi-mcp.com/internal/mcptools"
)

// NewMCPServer creates and returns an MCP server with all tools registered
func NewMCPServer() *server.MCPServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Register all tools
	s.AddTool(mcptools.NewCreateTodoMCPTool(), mcptools.CreateTodoHandler)
	s.AddTool(mcptools.NewDeleteTodoByIdMCPTool(), mcptools.DeleteTodoByIdHandler)
	s.AddTool(mcptools.NewGetTodoByIdMCPTool(), mcptools.GetTodoByIdHandler)
	s.AddTool(mcptools.NewListTodosMCPTool(), mcptools.ListTodosHandler)
	s.AddTool(mcptools.NewUpdateTodoByIdMCPTool(), mcptools.UpdateTodoByIdHandler)

	return s
}
