package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteProperty6 tool
const DeleteProperty6InputSchema = "{\n  \"properties\": {\n    \"propertyKey\": {\n      \"description\": \"The key of the user's property\",\n      \"type\": \"string\"\n    },\n    \"userKey\": {\n      \"description\": \"Key of the user whose property is to be removed\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"Username of the user whose property is to be removed\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteProperty6MCPTool creates the MCP Tool instance for DeleteProperty6
func NewDeleteProperty6MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProperty6",
		"Delete a specified user's property - Removes the property from the user identified by the key or by the id. The user who removes the property is required to have permissions to administer the user.",
		[]byte(DeleteProperty6InputSchema),
	)
}

// DeleteProperty6Handler is the handler function for the DeleteProperty6 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteProperty6Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/user/properties/{propertyKey}", args, []string{"propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteProperty6"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
