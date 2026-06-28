package mcptools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// ToolEntry pairs an MCP tool definition with its handler function.
type ToolEntry struct {
	Tool    mcp.Tool
	Handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// Registry maps tool names to their ToolEntry for dynamic tool discovery.
var Registry = map[string]ToolEntry{
	"GetPolicyViolations":                 {Tool: NewGetPolicyViolationsMCPTool(), Handler: GetPolicyViolationsHandler},
	"GetReportHistoryForApplication":      {Tool: NewGetReportHistoryForApplicationMCPTool(), Handler: GetReportHistoryForApplicationHandler},
	"GetSuggestedRemediationForComponent": {Tool: NewGetSuggestedRemediationForComponentMCPTool(), Handler: GetSuggestedRemediationForComponentHandler},
}
