package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the FullyUpdateProjectRole tool
const FullyUpdateProjectRoleInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"description\": {\n          \"example\": \"role description\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"MyRole\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"The role id\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the FullyUpdateProjectRole tool (Status: 200, Content-Type: application/json)
const FullyUpdateProjectRoleResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns updated role.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n  - **actors** (Type: array):\n    - **Items** (Type: object):\n      - **avatarUrl** (Type: string, uri):\n      - **name** (Type: string):\n          - Example: 'jira-developers'\n  - **description** (Type: string):\n      - Example: 'A project role that represents developers in a project'\n  - **id** (Type: integer, int64):\n      - Example: '10360'\n  - **name** (Type: string):\n      - Example: 'Developers'\n"

// NewFullyUpdateProjectRoleMCPTool creates the MCP Tool instance for FullyUpdateProjectRole
func NewFullyUpdateProjectRoleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FullyUpdateProjectRole",
		"Fully updates a role's name and description - Fully updates a roles. Both name and description must be given.",
		[]byte(FullyUpdateProjectRoleInputSchema),
	)
}

// FullyUpdateProjectRoleHandler is the handler function for the FullyUpdateProjectRole tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FullyUpdateProjectRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/role/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "FullyUpdateProjectRole"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
