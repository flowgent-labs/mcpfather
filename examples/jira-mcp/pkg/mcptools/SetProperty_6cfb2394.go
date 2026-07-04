package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/pkg/helpers"
	"time"
)

// Input Schema for the SetProperty_6cfb2394 tool
const SetProperty_6cfb2394InputSchema = "{\n  \"properties\": {\n    \"dashboardId\": {\n      \"description\": \"The dashboard id.\",\n      \"type\": \"string\"\n    },\n    \"itemId\": {\n      \"description\": \"The dashboard item on which the property will be set.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the dashboard item's property. The maximum length of the key is 255 bytes.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"dashboardId\",\n    \"itemId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty_6cfb2394MCPTool creates the MCP Tool instance for SetProperty_6cfb2394
func NewSetProperty_6cfb2394MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty_6cfb2394",
		"Set a property on a dashboard item - Sets the value of the property with a given key on the dashboard item identified by the id.",
		[]byte(SetProperty_6cfb2394InputSchema),
	)
}

// SetProperty_6cfb2394Handler is the handler function for the SetProperty_6cfb2394 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProperty_6cfb2394Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/dashboard/{dashboardId}/items/{itemId}/properties/{propertyKey}", args, []string{"dashboardId", "itemId", "propertyKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetProperty_6cfb2394")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
