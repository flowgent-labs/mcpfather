package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetProperty1_d12577d8 tool
const SetProperty1_d12577d8InputSchema = "{\n  \"properties\": {\n    \"commentId\": {\n      \"description\": \"the comment on which the property will be set.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"the key of the comment's property. The maximum length of the key is 255 bytes.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"commentId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty1_d12577d8MCPTool creates the MCP Tool instance for SetProperty1_d12577d8
func NewSetProperty1_d12577d8MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty1_d12577d8",
		"Set a property on a comment - Sets the value of the specified comment's property.",
		[]byte(SetProperty1_d12577d8InputSchema),
	)
}

// SetProperty1_d12577d8Handler is the handler function for the SetProperty1_d12577d8 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProperty1_d12577d8Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetProperty1_d12577d8"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
