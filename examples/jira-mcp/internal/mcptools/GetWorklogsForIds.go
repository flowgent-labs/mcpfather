package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetWorklogsForIds tool
const GetWorklogsForIdsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"a JSON object containing ids of worklogs to return\",\n      \"properties\": {\n        \"ids\": {\n          \"description\": \"List of worklog ids\",\n          \"example\": [\n            1,\n            2,\n            5,\n            10\n          ],\n          \"items\": {\n            \"description\": \"List of worklog ids\",\n            \"format\": \"int64\",\n            \"type\": \"integer\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetWorklogsForIds tool (Status: 200, Content-Type: application/json)
const GetWorklogsForIdsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the worklogs.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **created** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **id** (Type: string):\n      - Example: '100028'\n  - **timeSpentSeconds** (Type: integer, int64):\n      - Example: '12000'\n  - **updated** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **started** (Type: string):\n      - Example: '2010-07-14T18:23:23.733+0000'\n  - **updateAuthor** (Type: object):\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **avatarUrls** (Type: object):\n        - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n      - **Additional Properties**:\n        - **property value** (Type: string):\n            - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **key** (Type: string):\n        - Example: 'fred'\n    - **name** (Type: string):\n        - Example: 'Fred'\n    - **self** (Type: string):\n        - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n  - **visibility** (Type: object):\n    - **value** (Type: string):\n        - Example: 'jira-software-users'\n    - **type** (Type: string):\n        - Example: 'group'\n        - Enum: ['group', 'role']\n  - **[cyclic reference]**\n  - **comment** (Type: string):\n      - Example: 'I did some work here.'\n  - **timeSpent** (Type: string):\n      - Example: '3h 20m'\n  - **issueId** (Type: string):\n      - Example: '10002'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10010/worklog/10000'\n"

// NewGetWorklogsForIdsMCPTool creates the MCP Tool instance for GetWorklogsForIds
func NewGetWorklogsForIdsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetWorklogsForIds",
		"Returns worklogs for given ids. - Returns worklogs for given worklog ids. Only worklogs to which the calling user has permissions, will be included in the result. The returns set of worklogs is limited to 1000 elements.",
		[]byte(GetWorklogsForIdsInputSchema),
	)
}

// GetWorklogsForIdsHandler is the handler function for the GetWorklogsForIds tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetWorklogsForIdsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/worklog/list", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetWorklogsForIds")
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
