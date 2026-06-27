package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the RemoveCategory tool
const RemoveCategoryInputSchema = "{\n  \"properties\": {\n    \"categoryName\": {\n      \"description\": \"The name of the category to remove\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"The key of the space to remove the category from\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"categoryName\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewRemoveCategoryMCPTool creates the MCP Tool instance for RemoveCategory
func NewRemoveCategoryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RemoveCategory",
		"Remove a category from a space - Removes a category from a space, identified by spaceKey.\n\nExample request URI:\n"+"\x60"+"https://example.com/confluence/rest/api/space/TEST/category/example-category"+"\x60"+"",
		[]byte(RemoveCategoryInputSchema),
	)
}

// RemoveCategoryHandler is the handler function for the RemoveCategory tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RemoveCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/space/{spaceKey}/category/{categoryName}", args, []string{"categoryName", "spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RemoveCategory"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
