package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the DeleteTab tool
const DeleteTabInputSchema = "{\n  \"properties\": {\n    \"screenId\": {\n      \"description\": \"id of screen\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"tabId\": {\n      \"description\": \"id of tab\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"screenId\",\n    \"tabId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteTabMCPTool creates the MCP Tool instance for DeleteTab
func NewDeleteTabMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteTab",
		"Delete a tab from a screen - Deletes tab from given screen.",
		[]byte(DeleteTabInputSchema),
	)
}

// DeleteTabHandler is the handler function for the DeleteTab tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteTabHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/screens/{screenId}/tabs/{tabId}", args, []string{"screenId", "tabId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteTab")
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
