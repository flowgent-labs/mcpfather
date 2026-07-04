package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the AddUserToGroup tool
const AddUserToGroupInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"name\": {\n          \"example\": \"charlie\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"groupname\": {\n      \"description\": \"A name of requested group.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupname\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddUserToGroup tool (Status: 201, Content-Type: application/json)
const AddUserToGroupResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returns full representation of a Jira group in JSON format.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'jira-administrators'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators'\n  - **users** (Type: object):\n    - **items** (Type: array):\n        - Example: '[]'\n      - **Items** (Type: object):\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n        - **timeZone** (Type: string):\n            - Example: 'Australia/Sydney'\n        - **active** (Type: boolean):\n            - Example: 'true'\n        - **avatarUrls** (Type: object):\n            - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n          - **Additional Properties**:\n            - **property value** (Type: string):\n                - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n        - **displayName** (Type: string):\n            - Example: 'Fred F. User'\n        - **emailAddress** (Type: string):\n            - Example: 'fred@example.com'\n        - **key** (Type: string):\n            - Example: 'fred'\n        - **name** (Type: string):\n            - Example: 'Fred'\n    - **maxResults** (Type: integer, int32):\n        - Example: '50'\n    - **pagingCallback** (Type: object):\n    - **size** (Type: integer, int32):\n        - Example: '50'\n    - **backingListSize** (Type: integer, int32):\n    - **[cyclic reference]**\n"

// NewAddUserToGroupMCPTool creates the MCP Tool instance for AddUserToGroup
func NewAddUserToGroupMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddUserToGroup",
		"Add a user to a specified group - Adds given user to a group",
		[]byte(AddUserToGroupInputSchema),
	)
}

// AddUserToGroupHandler is the handler function for the AddUserToGroup tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddUserToGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/group/user", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddUserToGroup")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
