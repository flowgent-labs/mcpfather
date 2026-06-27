package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteProperty5 tool
const DeleteProperty5InputSchema = "{\n  \"properties\": {\n    \"projectIdOrKey\": {\n      \"description\": \"The project from which the property will be removed.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the property to remove.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectIdOrKey\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteProperty5MCPTool creates the MCP Tool instance for DeleteProperty5
func NewDeleteProperty5MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProperty5",
		"Delete property from project - Removes the property from the project identified by the key or by the id.",
		[]byte(DeleteProperty5InputSchema),
	)
}

// DeleteProperty5Handler is the handler function for the DeleteProperty5 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteProperty5Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/project/{projectIdOrKey}/properties/{propertyKey}", args, []string{"projectIdOrKey", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteProperty5"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
