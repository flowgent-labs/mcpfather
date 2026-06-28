package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the SetActors tool
const SetActorsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The actors to set for the role\",\n      \"properties\": {\n        \"categorisedActors\": {\n          \"additionalProperties\": {\n            \"example\": {\n              \"atlassian-group-role-actor\": [\n                \"jira-developers\"\n              ],\n              \"atlassian-user-role-actor\": [\n                \"admin\"\n              ]\n            },\n            \"items\": {\n              \"example\": \"{\\\"atlassian-user-role-actor\\\":[\\\"admin\\\"],\\\"atlassian-group-role-actor\\\":[\\\"jira-developers\\\"]}\",\n              \"type\": \"string\"\n            },\n            \"type\": \"array\"\n          },\n          \"example\": {\n            \"atlassian-group-role-actor\": [\n              \"jira-developers\"\n            ],\n            \"atlassian-user-role-actor\": [\n              \"admin\"\n            ]\n          },\n          \"type\": \"object\"\n        },\n        \"id\": {\n          \"example\": 10360,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"The project role id\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"The project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"id\",\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the SetActors tool (Status: 200, Content-Type: application/json)
const SetActorsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Role details and its actors after modification\n\n## Response Structure\n\n- Structure (Type: object):\n  - **actors** (Type: array):\n    - **Items** (Type: object):\n      - **avatarUrl** (Type: string, uri):\n      - **name** (Type: string):\n          - Example: 'jira-developers'\n  - **description** (Type: string):\n      - Example: 'A project role that represents developers in a project'\n  - **id** (Type: integer, int64):\n      - Example: '10360'\n  - **name** (Type: string):\n      - Example: 'Developers'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n"

// NewSetActorsMCPTool creates the MCP Tool instance for SetActors
func NewSetActorsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetActors",
		"Update project role with actors - Updates a project role to include the specified actors (users or groups). Can be also used to clear roles to not include any users or groups. For user actors, their usernames should be used.",
		[]byte(SetActorsInputSchema),
	)
}

// SetActorsHandler is the handler function for the SetActors tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetActorsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/project/{projectIdOrKey}/role/{id}", args, []string{"id", "projectIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetActors")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
