package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the AddWorklog tool
const AddWorklogInputSchema = "{\n  \"properties\": {\n    \"adjustEstimate\": {\n      \"description\": \"Allows you to provide specific instructions to update the remaining time estimate of the issue. Valid values are: new, leave, manual, auto\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"Worklog create request\",\n      \"properties\": {\n        \"author\": {\n          \"properties\": {\n            \"active\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"avatarUrls\": {\n              \"additionalProperties\": {\n                \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n                \"type\": \"string\"\n              },\n              \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n              \"type\": \"object\"\n            },\n            \"displayName\": {\n              \"example\": \"Fred F. User\",\n              \"type\": \"string\"\n            },\n            \"emailAddress\": {\n              \"example\": \"fred@example.com\",\n              \"type\": \"string\"\n            },\n            \"key\": {\n              \"example\": \"fred\",\n              \"type\": \"string\"\n            },\n            \"name\": {\n              \"example\": \"Fred\",\n              \"type\": \"string\"\n            },\n            \"self\": {\n              \"example\": \"http://www.example.com/jira/rest/api/2/user?username=fred\",\n              \"type\": \"string\"\n            },\n            \"timeZone\": {\n              \"example\": \"Australia/Sydney\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"comment\": {\n          \"example\": \"I did some work here.\",\n          \"type\": \"string\"\n        },\n        \"created\": {\n          \"example\": \"2010-07-14T18:23:23.733+0000\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"100028\",\n          \"type\": \"string\"\n        },\n        \"issueId\": {\n          \"example\": \"10002\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/issue/10010/worklog/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"started\": {\n          \"example\": \"2010-07-14T18:23:23.733+0000\",\n          \"type\": \"string\"\n        },\n        \"timeSpent\": {\n          \"example\": \"3h 20m\",\n          \"type\": \"string\"\n        },\n        \"timeSpentSeconds\": {\n          \"example\": 12000,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"updateAuthor\": {},\n        \"updated\": {\n          \"example\": \"2010-07-14T18:23:23.733+0000\",\n          \"type\": \"string\"\n        },\n        \"visibility\": {\n          \"properties\": {\n            \"type\": {\n              \"enum\": [\n                \"group\",\n                \"role\"\n              ],\n              \"example\": \"group\",\n              \"type\": \"string\"\n            },\n            \"value\": {\n              \"example\": \"jira-software-users\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"a string containing the issue id or key the worklog will be added to\",\n      \"type\": \"string\"\n    },\n    \"newEstimate\": {\n      \"description\": \"Required when 'new' is selected for adjustEstimate. e.g. \\\"2d\\\"\",\n      \"type\": \"string\"\n    },\n    \"reduceBy\": {\n      \"description\": \"Required when 'manual' is selected for adjustEstimate. e.g. \\\"2d\\\"\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddWorklog tool (Status: 201, Content-Type: application/json)
const AddWorklogResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returned if add was successful.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **timeSpentSeconds** (Type: integer, int64):\n      - Example: '12000'\n  - **author** (Type: object):\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **avatarUrls** (Type: object):\n        - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n      - **Additional Properties**:\n        - **property value** (Type: string):\n            - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **key** (Type: string):\n        - Example: 'fred'\n    - **name** (Type: string):\n        - Example: 'Fred'\n    - **self** (Type: string):\n        - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n  - **issueId** (Type: string):\n      - Example: '10002'\n  - **created** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10010/worklog/10000'\n  - **timeSpent** (Type: string):\n      - Example: '3h 20m'\n  - **comment** (Type: string):\n      - Example: 'I did some work here.'\n  - **started** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **[cyclic reference]**\n  - **updated** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **visibility** (Type: object):\n    - **value** (Type: string):\n        - Example: 'jira-software-users'\n    - **type** (Type: string):\n        - Example: 'group'\n        - Enum: ['group', 'role']\n  - **id** (Type: string):\n      - Example: '100028'\n"

// NewAddWorklogMCPTool creates the MCP Tool instance for AddWorklog
func NewAddWorklogMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddWorklog",
		"Add a worklog entry - Adds a new worklog entry to an issue.",
		[]byte(AddWorklogInputSchema),
	)
}

// AddWorklogHandler is the handler function for the AddWorklog tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddWorklogHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issue/{issueIdOrKey}/worklog", args, []string{"issueIdOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddWorklog")
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
