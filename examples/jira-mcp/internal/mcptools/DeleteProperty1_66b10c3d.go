package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the DeleteProperty1_66b10c3d tool
const DeleteProperty1_66b10c3dInputSchema = "{\n  \"properties\": {\n    \"dashboardId\": {\n      \"description\": \"The dashboard id.\",\n      \"type\": \"string\"\n    },\n    \"itemId\": {\n      \"description\": \"The dashboard item from which the property will be removed.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the property to remove.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"dashboardId\",\n    \"itemId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteProperty1_66b10c3dMCPTool creates the MCP Tool instance for DeleteProperty1_66b10c3d
func NewDeleteProperty1_66b10c3dMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProperty1_66b10c3d",
		"Delete a property from a dashboard item - Removes the property from the dashboard item identified by the key or by the id.",
		[]byte(DeleteProperty1_66b10c3dInputSchema),
	)
}

// DeleteProperty1_66b10c3dHandler is the handler function for the DeleteProperty1_66b10c3d tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteProperty1_66b10c3dHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteProperty1_66b10c3d")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
