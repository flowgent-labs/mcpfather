package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteProperty tool
const DeletePropertyInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"boardId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeletePropertyMCPTool creates the MCP Tool instance for DeleteProperty
func NewDeletePropertyMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProperty",
		"Delete a property from a board - Removes the property from the board identified by the id. Ths user removing the property is required to have permissions to modify the board.",
		[]byte(DeletePropertyInputSchema),
	)
}

// DeletePropertyHandler is the handler function for the DeleteProperty tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeletePropertyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/agile/1.0/board/{boardId}/properties/{propertyKey}", args, []string{"boardId", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteProperty"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
