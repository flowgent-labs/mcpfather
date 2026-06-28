package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the SetProperty1_9fc616f7 tool
const SetProperty1_9fc616f7InputSchema = "{\n  \"properties\": {\n    \"commentId\": {\n      \"description\": \"the comment on which the property will be set.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"the key of the comment's property. The maximum length of the key is 255 bytes.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"commentId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty1_9fc616f7MCPTool creates the MCP Tool instance for SetProperty1_9fc616f7
func NewSetProperty1_9fc616f7MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty1_9fc616f7",
		"Set a property on a comment - Sets the value of the specified comment's property.",
		[]byte(SetProperty1_9fc616f7InputSchema),
	)
}

// SetProperty1_9fc616f7Handler is the handler function for the SetProperty1_9fc616f7 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProperty1_9fc616f7Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/comment/{commentId}/properties/{propertyKey}", args, []string{"commentId", "propertyKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetProperty1_9fc616f7")
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
