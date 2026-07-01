package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetComment tool
const GetCommentInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Optional flags: renderedBody (provides body rendered in HTML)\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"Comment id\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetComment tool (Status: 200, Content-Type: application/json)
const GetCommentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full representation of a Jira comment in JSON format.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **renderedBody** (Type: string):\n      - Example: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.'\n  - **updateAuthor** (Type: object):\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **avatarUrls** (Type: object):\n        - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n      - **Additional Properties**:\n        - **property value** (Type: string):\n            - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **key** (Type: string):\n        - Example: 'fred'\n    - **name** (Type: string):\n        - Example: 'Fred'\n    - **self** (Type: string):\n        - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n  - **updated** (Type: string):\n      - Example: '2012-07-06T18:30:00.000+0000'\n  - **visibility** (Type: object):\n    - **value** (Type: string):\n        - Example: 'jira-software-users'\n    - **type** (Type: string):\n        - Example: 'group'\n        - Enum: ['group', 'role']\n  - **id** (Type: string):\n      - Example: '10000'\n  - **body** (Type: string):\n      - Example: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.'\n  - **properties** (Type: array):\n    - **Items** (Type: object):\n      - **key** (Type: string):\n          - Example: 'issue.support'\n      - **value** (Type: string):\n          - Example: '{\"hipchat.room.id\":\"support-123\",\"support.time\":\"1m\"}'\n  - **[cyclic reference]**\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10010/comment/10000'\n  - **created** (Type: string):\n      - Example: '2012-07-06T18:30:00.000+0000'\n"

// NewGetCommentMCPTool creates the MCP Tool instance for GetComment
func NewGetCommentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetComment",
		"Get a comment by id - Returns a single comment.",
		[]byte(GetCommentInputSchema),
	)
}

// GetCommentHandler is the handler function for the GetComment tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issue/{issueIdOrKey}/comment/{id}", args, []string{"id", "issueIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetComment")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
