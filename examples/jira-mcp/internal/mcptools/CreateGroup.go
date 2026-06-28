package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateGroup tool
const CreateGroupInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"name\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the CreateGroup tool (Status: 201, Content-Type: application/json)
const CreateGroupResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returns full representation of a Jira group in JSON format.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'jira-administrators'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators'\n  - **users** (Type: object):\n    - **backingListSize** (Type: integer, int32):\n    - **callback** (Type: object):\n    - **items** (Type: array):\n        - Example: '[]'\n      - **Items** (Type: object):\n        - **key** (Type: string):\n            - Example: 'fred'\n        - **name** (Type: string):\n            - Example: 'Fred'\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n        - **timeZone** (Type: string):\n            - Example: 'Australia/Sydney'\n        - **active** (Type: boolean):\n            - Example: 'true'\n        - **avatarUrls** (Type: object):\n            - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n          - **Additional Properties**:\n            - **property value** (Type: string):\n                - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n        - **displayName** (Type: string):\n            - Example: 'Fred F. User'\n        - **emailAddress** (Type: string):\n            - Example: 'fred@example.com'\n    - **maxResults** (Type: integer, int32):\n        - Example: '50'\n    - **[cyclic reference]**\n    - **size** (Type: integer, int32):\n        - Example: '50'\n"

// NewCreateGroupMCPTool creates the MCP Tool instance for CreateGroup
func NewCreateGroupMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateGroup",
		"Create a group with given parameters - Creates a group by given group parameter",
		[]byte(CreateGroupInputSchema),
	)
}

// CreateGroupHandler is the handler function for the CreateGroup tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/group", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateGroup")
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
