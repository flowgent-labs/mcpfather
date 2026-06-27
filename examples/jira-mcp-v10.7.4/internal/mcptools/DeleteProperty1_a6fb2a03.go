package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteProperty1_a6fb2a03 tool
const DeleteProperty1_a6fb2a03InputSchema = "{\n  \"properties\": {\n    \"dashboardId\": {\n      \"description\": \"The dashboard id.\",\n      \"type\": \"string\"\n    },\n    \"itemId\": {\n      \"description\": \"The dashboard item from which the property will be removed.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the property to remove.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"dashboardId\",\n    \"itemId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteProperty1_a6fb2a03MCPTool creates the MCP Tool instance for DeleteProperty1_a6fb2a03
func NewDeleteProperty1_a6fb2a03MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProperty1_a6fb2a03",
		"Delete a property from a dashboard item - Removes the property from the dashboard item identified by the key or by the id.",
		[]byte(DeleteProperty1_a6fb2a03InputSchema),
	)
}

// DeleteProperty1_a6fb2a03Handler is the handler function for the DeleteProperty1_a6fb2a03 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteProperty1_a6fb2a03Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/dashboard/{dashboardId}/items/{itemId}/properties/{propertyKey}", args, []string{"dashboardId", "itemId", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteProperty1_a6fb2a03"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
