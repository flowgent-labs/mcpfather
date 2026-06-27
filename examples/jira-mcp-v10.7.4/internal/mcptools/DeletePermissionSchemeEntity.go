package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeletePermissionSchemeEntity tool
const DeletePermissionSchemeEntityInputSchema = "{\n  \"properties\": {\n    \"permissionId\": {\n      \"description\": \"The id of the permission grant.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"schemeId\": {\n      \"description\": \"The id of the permission scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"permissionId\",\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeletePermissionSchemeEntityMCPTool creates the MCP Tool instance for DeletePermissionSchemeEntity
func NewDeletePermissionSchemeEntityMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeletePermissionSchemeEntity",
		"Delete a permission grant from a scheme - Deletes a permission grant from a permission scheme.",
		[]byte(DeletePermissionSchemeEntityInputSchema),
	)
}

// DeletePermissionSchemeEntityHandler is the handler function for the DeletePermissionSchemeEntity tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeletePermissionSchemeEntityHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/permissionscheme/{schemeId}/permission/{permissionId}", args, []string{"permissionId", "schemeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeletePermissionSchemeEntity"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
