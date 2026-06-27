package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProjectRole tool
const GetProjectRoleInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The project role id\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"The project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProjectRole tool (Status: 200, Content-Type: application/json)
const GetProjectRoleResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Role details and its actors\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n      - Example: 'A project role that represents developers in a project'\n  - **id** (Type: integer, int64):\n      - Example: '10360'\n  - **name** (Type: string):\n      - Example: 'Developers'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n  - **actors** (Type: array):\n    - **Items** (Type: object):\n      - **avatarUrl** (Type: string, uri):\n      - **name** (Type: string):\n          - Example: 'jira-developers'\n"

// NewGetProjectRoleMCPTool creates the MCP Tool instance for GetProjectRole
func NewGetProjectRoleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProjectRole",
		"Get details for a project role - Returns the details for a given project role in a project.",
		[]byte(GetProjectRoleInputSchema),
	)
}

// GetProjectRoleHandler is the handler function for the GetProjectRole tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProjectRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/project/{projectIdOrKey}/role/{id}", args, []string{"id", "projectIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProjectRole"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
