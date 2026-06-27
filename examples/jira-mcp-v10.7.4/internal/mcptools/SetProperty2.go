package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetProperty2 tool
const SetProperty2InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The value of the issue's property\",\n      \"format\": \"json\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the issue's property\",\n      \"maxLength\": 255,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"issueIdOrKey\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty2MCPTool creates the MCP Tool instance for SetProperty2
func NewSetProperty2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty2",
		"Update the value of a specific issue's property - Sets the value of the specified issue's property.",
		[]byte(SetProperty2InputSchema),
	)
}

// SetProperty2Handler is the handler function for the SetProperty2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProperty2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issue/{issueIdOrKey}/properties/{propertyKey}", args, []string{"issueIdOrKey", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetProperty2"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
