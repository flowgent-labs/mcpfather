package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the RemoveProprietaryComponentNames tool
const RemoveProprietaryComponentNamesInputSchema = "{\n  \"properties\": {\n    \"format\": {\n      \"description\": \"Format for which the proprietary namespaces are being removed.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"format\"\n  ],\n  \"type\": \"object\"\n}"

// NewRemoveProprietaryComponentNamesMCPTool creates the MCP Tool instance for RemoveProprietaryComponentNames
func NewRemoveProprietaryComponentNamesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RemoveProprietaryComponentNames",
		"Removes proprietary component namespaces for the specified format.\n\nPermissions required: Evaluate Individual Components",
		[]byte(RemoveProprietaryComponentNamesInputSchema),
	)
}

// RemoveProprietaryComponentNamesHandler is the handler function for the RemoveProprietaryComponentNames tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RemoveProprietaryComponentNamesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/firewall/namespace_confusion/{format}", args, []string{"format"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RemoveProprietaryComponentNames"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
