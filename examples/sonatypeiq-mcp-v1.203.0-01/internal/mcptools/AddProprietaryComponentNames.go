package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddProprietaryComponentNames tool
const AddProprietaryComponentNamesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"List of namespaces to register as proprietary for this format.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"format\": {\n      \"description\": \"Format for which the proprietary namespaces are being added.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"format\"\n  ],\n  \"type\": \"object\"\n}"

// NewAddProprietaryComponentNamesMCPTool creates the MCP Tool instance for AddProprietaryComponentNames
func NewAddProprietaryComponentNamesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddProprietaryComponentNames",
		"Adds a list of proprietary component namespaces for the specified format to prevent namespace confusion attacks.\n\nPermissions required: Evaluate Individual Components",
		[]byte(AddProprietaryComponentNamesInputSchema),
	)
}

// AddProprietaryComponentNamesHandler is the handler function for the AddProprietaryComponentNames tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddProprietaryComponentNamesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/firewall/namespace_confusion/{format}", args, []string{"format"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddProprietaryComponentNames")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
