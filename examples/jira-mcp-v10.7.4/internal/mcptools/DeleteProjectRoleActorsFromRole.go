package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteProjectRoleActorsFromRole tool
const DeleteProjectRoleActorsFromRoleInputSchema = "{\n  \"properties\": {\n    \"group\": {\n      \"description\": \"If given, removes an actor from given role\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"The role id to remove the actors from\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"user\": {\n      \"description\": \"If given, removes an actor from given role\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the DeleteProjectRoleActorsFromRole tool (Status: 200, Content-Type: application/json)
const DeleteProjectRoleActorsFromRoleResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns updated actors list.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **actors** (Type: array):\n    - **Items** (Type: object):\n      - **avatarUrl** (Type: string, uri):\n      - **name** (Type: string):\n          - Example: 'jira-developers'\n"

// NewDeleteProjectRoleActorsFromRoleMCPTool creates the MCP Tool instance for DeleteProjectRoleActorsFromRole
func NewDeleteProjectRoleActorsFromRoleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProjectRoleActorsFromRole",
		"Removes default actor from a role - Removes default actor from the given role.",
		[]byte(DeleteProjectRoleActorsFromRoleInputSchema),
	)
}

// DeleteProjectRoleActorsFromRoleHandler is the handler function for the DeleteProjectRoleActorsFromRole tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteProjectRoleActorsFromRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/role/{id}/actors", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteProjectRoleActorsFromRole"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
