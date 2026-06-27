package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteProjectRole tool
const DeleteProjectRoleInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The role id\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"swap\": {\n      \"description\": \"If given, removes a role even if it is used in scheme by replacing the role with the given one\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteProjectRoleMCPTool creates the MCP Tool instance for DeleteProjectRole
func NewDeleteProjectRoleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProjectRole",
		"Deletes a role - Deletes a role. May return 403 in the future",
		[]byte(DeleteProjectRoleInputSchema),
	)
}

// DeleteProjectRoleHandler is the handler function for the DeleteProjectRole tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteProjectRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/role/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteProjectRole"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
