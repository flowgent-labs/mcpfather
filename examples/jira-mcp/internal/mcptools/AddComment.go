package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the AddComment tool
const AddCommentInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Comment create request\",\n      \"properties\": {\n        \"author\": {},\n        \"body\": {\n          \"example\": \"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.\",\n          \"type\": \"string\"\n        },\n        \"created\": {\n          \"example\": \"2012-07-06T18:30:00.000+0000\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"properties\": {\n          \"items\": {\n            \"properties\": {\n              \"key\": {\n                \"example\": \"issue.support\",\n                \"type\": \"string\"\n              },\n              \"value\": {\n                \"example\": \"{\\\"hipchat.room.id\\\":\\\"support-123\\\",\\\"support.time\\\":\\\"1m\\\"}\",\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        },\n        \"renderedBody\": {\n          \"example\": \"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/issue/10010/comment/10000\",\n          \"type\": \"string\"\n        },\n        \"updateAuthor\": {\n          \"properties\": {\n            \"active\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"avatarUrls\": {\n              \"additionalProperties\": {\n                \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n                \"type\": \"string\"\n              },\n              \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n              \"type\": \"object\"\n            },\n            \"displayName\": {\n              \"example\": \"Fred F. User\",\n              \"type\": \"string\"\n            },\n            \"emailAddress\": {\n              \"example\": \"fred@example.com\",\n              \"type\": \"string\"\n            },\n            \"key\": {\n              \"example\": \"fred\",\n              \"type\": \"string\"\n            },\n            \"name\": {\n              \"example\": \"Fred\",\n              \"type\": \"string\"\n            },\n            \"self\": {\n              \"example\": \"http://www.example.com/jira/rest/api/2/user?username=fred\",\n              \"type\": \"string\"\n            },\n            \"timeZone\": {\n              \"example\": \"Australia/Sydney\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"updated\": {\n          \"example\": \"2012-07-06T18:30:00.000+0000\",\n          \"type\": \"string\"\n        },\n        \"visibility\": {\n          \"properties\": {\n            \"type\": {\n              \"enum\": [\n                \"group\",\n                \"role\"\n              ],\n              \"example\": \"group\",\n              \"type\": \"string\"\n            },\n            \"value\": {\n              \"example\": \"jira-software-users\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"expand\": {\n      \"description\": \"Optional flags: renderedBody (provides body rendered in HTML)\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddComment tool (Status: 201, Content-Type: application/json)
const AddCommentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returned if add was successful.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **renderedBody** (Type: string):\n      - Example: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.'\n  - **updateAuthor** (Type: object):\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **avatarUrls** (Type: object):\n        - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n      - **Additional Properties**:\n        - **property value** (Type: string):\n            - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **key** (Type: string):\n        - Example: 'fred'\n    - **name** (Type: string):\n        - Example: 'Fred'\n    - **self** (Type: string):\n        - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n  - **updated** (Type: string):\n      - Example: '2012-07-06T18:30:00.000+0000'\n  - **visibility** (Type: object):\n    - **value** (Type: string):\n        - Example: 'jira-software-users'\n    - **type** (Type: string):\n        - Example: 'group'\n        - Enum: ['group', 'role']\n  - **id** (Type: string):\n      - Example: '10000'\n  - **body** (Type: string):\n      - Example: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.'\n  - **properties** (Type: array):\n    - **Items** (Type: object):\n      - **key** (Type: string):\n          - Example: 'issue.support'\n      - **value** (Type: string):\n          - Example: '{\"hipchat.room.id\":\"support-123\",\"support.time\":\"1m\"}'\n  - **[cyclic reference]**\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10010/comment/10000'\n  - **created** (Type: string):\n      - Example: '2012-07-06T18:30:00.000+0000'\n"

// NewAddCommentMCPTool creates the MCP Tool instance for AddComment
func NewAddCommentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddComment",
		"Add a comment - Adds a new comment to an issue.",
		[]byte(AddCommentInputSchema),
	)
}

// AddCommentHandler is the handler function for the AddComment tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issue/{issueIdOrKey}/comment", args, []string{"issueIdOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddComment")
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
