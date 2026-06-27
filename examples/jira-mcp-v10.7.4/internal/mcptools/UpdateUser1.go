package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateUser1 tool
const UpdateUser1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"User details\",\n      \"properties\": {\n        \"active\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"applicationKeys\": {\n          \"example\": [\n            \"jira-core\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"jira-core\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"displayName\": {\n          \"example\": \"Charlie of Atlassian\",\n          \"type\": \"string\"\n        },\n        \"emailAddress\": {\n          \"example\": \"charlie@atlassian.com\",\n          \"type\": \"string\"\n        },\n        \"key\": {\n          \"example\": \"charlie\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"charlie\",\n          \"type\": \"string\"\n        },\n        \"notification\": {\n          \"example\": \"HTML\",\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"example\": \"abracadabra\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/user?username=charlie\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"key\": {\n      \"description\": \"user key\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"the username\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateUser1 tool (Status: 200, Content-Type: application/json)
const UpdateUser1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the user exists and the caller has permission to edit it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **displayName** (Type: string):\n      - Example: 'Charlie of Atlassian'\n  - **emailAddress** (Type: string):\n      - Example: 'charlie@atlassian.com'\n  - **active** (Type: boolean):\n      - Example: 'true'\n  - **applicationKeys** (Type: array):\n      - Example: '[\"jira-core\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-core\"]'\n  - **name** (Type: string):\n      - Example: 'charlie'\n  - **password** (Type: string):\n      - Example: 'abracadabra'\n  - **key** (Type: string):\n      - Example: 'charlie'\n  - **notification** (Type: string):\n      - Example: 'HTML'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/user?username=charlie'\n"

// NewUpdateUser1MCPTool creates the MCP Tool instance for UpdateUser1
func NewUpdateUser1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateUser1",
		"Update user details - Modify user. The 'value' fields present will override the existing value. Fields skipped in request will not be changed.",
		[]byte(UpdateUser1InputSchema),
	)
}

// UpdateUser1Handler is the handler function for the UpdateUser1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateUser1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/user", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateUser1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
