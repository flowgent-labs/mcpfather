package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the UpdateWorklog tool
const UpdateWorklogInputSchema = "{\n  \"properties\": {\n    \"adjustEstimate\": {\n      \"description\": \"allows you to provide specific instructions to update the remaining time estimate of the issue. Valid values are: new, leave, auto\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"Worklog update request\",\n      \"properties\": {\n        \"author\": {},\n        \"comment\": {\n          \"example\": \"I did some work here.\",\n          \"type\": \"string\"\n        },\n        \"created\": {\n          \"example\": \"2010-07-14T18:23:23.733+0000\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"100028\",\n          \"type\": \"string\"\n        },\n        \"issueId\": {\n          \"example\": \"10002\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/issue/10010/worklog/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"started\": {\n          \"example\": \"2010-07-14T18:23:23.733+0000\",\n          \"type\": \"string\"\n        },\n        \"timeSpent\": {\n          \"example\": \"3h 20m\",\n          \"type\": \"string\"\n        },\n        \"timeSpentSeconds\": {\n          \"example\": 12000,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"updateAuthor\": {\n          \"properties\": {\n            \"active\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"avatarUrls\": {\n              \"additionalProperties\": {\n                \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n                \"type\": \"string\"\n              },\n              \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n              \"type\": \"object\"\n            },\n            \"displayName\": {\n              \"example\": \"Fred F. User\",\n              \"type\": \"string\"\n            },\n            \"emailAddress\": {\n              \"example\": \"fred@example.com\",\n              \"type\": \"string\"\n            },\n            \"key\": {\n              \"example\": \"fred\",\n              \"type\": \"string\"\n            },\n            \"name\": {\n              \"example\": \"Fred\",\n              \"type\": \"string\"\n            },\n            \"self\": {\n              \"example\": \"http://www.example.com/jira/rest/api/2/user?username=fred\",\n              \"type\": \"string\"\n            },\n            \"timeZone\": {\n              \"example\": \"Australia/Sydney\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"updated\": {\n          \"example\": \"2010-07-14T18:23:23.733+0000\",\n          \"type\": \"string\"\n        },\n        \"visibility\": {\n          \"properties\": {\n            \"type\": {\n              \"enum\": [\n                \"group\",\n                \"role\"\n              ],\n              \"example\": \"group\",\n              \"type\": \"string\"\n            },\n            \"value\": {\n              \"example\": \"jira-software-users\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"id of the worklog to be updated\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"a string containing the issue id or key the worklog belongs to\",\n      \"type\": \"string\"\n    },\n    \"newEstimate\": {\n      \"description\": \"required when 'new' is selected for adjustEstimate\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateWorklog tool (Status: 200, Content-Type: application/json)
const UpdateWorklogResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if update was successful.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **created** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **id** (Type: string):\n      - Example: '100028'\n  - **timeSpentSeconds** (Type: integer, int64):\n      - Example: '12000'\n  - **updated** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **started** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **updateAuthor** (Type: object):\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **key** (Type: string):\n        - Example: 'fred'\n    - **name** (Type: string):\n        - Example: 'Fred'\n    - **self** (Type: string):\n        - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **avatarUrls** (Type: object):\n        - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n      - **Additional Properties**:\n        - **property value** (Type: string):\n            - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n  - **visibility** (Type: object):\n    - **value** (Type: string):\n        - Example: 'jira-software-users'\n    - **type** (Type: string):\n        - Example: 'group'\n        - Enum: ['group', 'role']\n  - **[cyclic reference]**\n  - **comment** (Type: string):\n      - Example: 'I did some work here.'\n  - **timeSpent** (Type: string):\n      - Example: '3h 20m'\n  - **issueId** (Type: string):\n      - Example: '10002'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10010/worklog/10000'\n"

// NewUpdateWorklogMCPTool creates the MCP Tool instance for UpdateWorklog
func NewUpdateWorklogMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateWorklog",
		"Update a worklog entry - Updates an existing worklog entry. Note that:\n- Fields possible for editing are: comment, visibility, started, timeSpent and timeSpentSeconds.\n- Either timeSpent or timeSpentSeconds can be set.\n- Fields which are not set will not be updated.\n- For a request to be valid, it has to have at least one field change.",
		[]byte(UpdateWorklogInputSchema),
	)
}

// UpdateWorklogHandler is the handler function for the UpdateWorklog tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateWorklogHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issue/{issueIdOrKey}/worklog/{id}", args, []string{"id", "issueIdOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateWorklog")
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
