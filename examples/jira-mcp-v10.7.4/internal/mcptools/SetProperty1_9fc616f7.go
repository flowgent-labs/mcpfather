package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetProperty1_9fc616f7 tool
const SetProperty1_9fc616f7InputSchema = "{\n  \"properties\": {\n    \"propertyKey\": {\n      \"type\": \"string\"\n    },\n    \"sprintId\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"propertyKey\",\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty1_9fc616f7MCPTool creates the MCP Tool instance for SetProperty1_9fc616f7
func NewSetProperty1_9fc616f7MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty1_9fc616f7",
		"Update a sprint's property - Sets the value of the specified sprint's property. You can use this resource to store a custom data against the sprint identified by the id. The user who stores the data is required to have permissions to modify the sprint.",
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
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/agile/1.0/sprint/{sprintId}/properties/{propertyKey}", args, []string{"propertyKey", "sprintId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetProperty1_9fc616f7"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
