package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the RemoveField tool
const RemoveFieldInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"id of field\",\n      \"type\": \"string\"\n    },\n    \"screenId\": {\n      \"description\": \"id of screen\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"tabId\": {\n      \"description\": \"id of tab\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"screenId\",\n    \"tabId\"\n  ],\n  \"type\": \"object\"\n}"

// NewRemoveFieldMCPTool creates the MCP Tool instance for RemoveField
func NewRemoveFieldMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RemoveField",
		"Remove field from tab - Removes field from given tab.",
		[]byte(RemoveFieldInputSchema),
	)
}

// RemoveFieldHandler is the handler function for the RemoveField tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RemoveFieldHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/screens/{screenId}/tabs/{tabId}/fields/{id}", args, []string{"id", "screenId", "tabId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RemoveField"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
