package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the UpdateUser tool
const UpdateUserInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The new user details to be set.\",\n      \"properties\": {\n        \"active\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"applicationKeys\": {\n          \"example\": [\n            \"jira-core\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"jira-core\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"displayName\": {\n          \"example\": \"Charlie of Atlassian\",\n          \"type\": \"string\"\n        },\n        \"emailAddress\": {\n          \"example\": \"charlie@atlassian.com\",\n          \"type\": \"string\"\n        },\n        \"key\": {\n          \"example\": \"charlie\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"charlie\",\n          \"type\": \"string\"\n        },\n        \"notification\": {\n          \"example\": \"HTML\",\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"example\": \"abracadabra\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/user?username=charlie\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateUser tool (Status: 200, Content-Type: application/json)
const UpdateUserResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Confirmation that the user was updated.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **active** (Type: boolean):\n      - Example: 'true'\n  - **name** (Type: string):\n      - Example: 'charlie'\n  - **notification** (Type: string):\n      - Example: 'HTML'\n  - **password** (Type: string):\n      - Example: 'abracadabra'\n  - **displayName** (Type: string):\n      - Example: 'Charlie of Atlassian'\n  - **applicationKeys** (Type: array):\n      - Example: '[\"jira-core\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-core\"]'\n  - **emailAddress** (Type: string):\n      - Example: 'charlie@atlassian.com'\n  - **key** (Type: string):\n      - Example: 'charlie'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/user?username=charlie'\n"

// NewUpdateUserMCPTool creates the MCP Tool instance for UpdateUser
func NewUpdateUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateUser",
		"Update currently logged user - Modify currently logged user. The 'value' fields present will override the existing value. Fields skipped in request will not be changed. Only email and display name can be change that way. Requires user password.",
		[]byte(UpdateUserInputSchema),
	)
}

// UpdateUserHandler is the handler function for the UpdateUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/myself", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateUser")
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
