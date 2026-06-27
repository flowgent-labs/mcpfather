package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteRepositoryManager tool
const DeleteRepositoryManagerInputSchema = "{\n  \"properties\": {\n    \"repositoryManagerId\": {\n      \"description\": \"Enter the repository manager ID.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"repositoryManagerId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteRepositoryManagerMCPTool creates the MCP Tool instance for DeleteRepositoryManager
func NewDeleteRepositoryManagerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteRepositoryManager",
		"Use this method to delete an existing repository manager.\n\nPermissions required: Edit IQ Elements",
		[]byte(DeleteRepositoryManagerInputSchema),
	)
}

// DeleteRepositoryManagerHandler is the handler function for the DeleteRepositoryManager tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteRepositoryManagerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/firewall/repositoryManagers/{repositoryManagerId}", args, []string{"repositoryManagerId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteRepositoryManager"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
