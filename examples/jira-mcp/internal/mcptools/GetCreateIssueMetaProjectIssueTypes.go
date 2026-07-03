package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetCreateIssueMetaProjectIssueTypes tool
const GetCreateIssueMetaProjectIssueTypesInputSchema = "{\n  \"properties\": {\n    \"maxResults\": {\n      \"description\": \"How many results on the page should be included\",\n      \"type\": \"string\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"Project id or key\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"description\": \"The page offset\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetCreateIssueMetaProjectIssueTypes tool (Status: 200, Content-Type: application/json)
const GetCreateIssueMetaProjectIssueTypesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the metadata for issue types used for creating issues.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **avatarId** (Type: integer, int64):\n      - Example: '10002'\n  - **description** (Type: string):\n      - Example: 'A problem which impairs or prevents the functions of the product.'\n  - **fields** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: object):\n        - **allowedValues** (Type: array):\n            - Example: '[\"red\",\"blue\",\"default value\"]'\n          - **Items** (Type: object):\n              - Example: '[\"red\",\"blue\",\"default value\"]'\n        - **hasDefaultValue** (Type: boolean):\n            - Example: 'true'\n        - **operations** (Type: array):\n            - Example: '[\"set\",\"add\"]'\n          - **Items** (Type: string):\n              - Example: '[\"set\",\"add\"]'\n        - **name** (Type: string):\n            - Example: 'My Multi Select'\n        - **autoCompleteUrl** (Type: string):\n            - Example: '/rest/api/2/customFieldOption/10000'\n        - **defaultValue** (Type: object):\n        - **required** (Type: boolean):\n            - Example: 'true'\n        - **fieldId** (Type: string):\n            - Example: 'customfield_10000'\n        - **schema** (Type: object):\n            - Example: '{}'\n          - **custom** (Type: string):\n              - Example: 'null'\n          - **customId** (Type: integer, int64):\n          - **items** (Type: string):\n              - Example: 'null'\n          - **system** (Type: string):\n              - Example: 'summary'\n          - **type** (Type: string):\n              - Example: 'string'\n  - **iconUrl** (Type: string):\n      - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n  - **id** (Type: string):\n      - Example: '1'\n  - **name** (Type: string):\n      - Example: 'Bug'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n  - **subtask** (Type: boolean):\n      - Example: 'false'\n"

// NewGetCreateIssueMetaProjectIssueTypesMCPTool creates the MCP Tool instance for GetCreateIssueMetaProjectIssueTypes
func NewGetCreateIssueMetaProjectIssueTypesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCreateIssueMetaProjectIssueTypes",
		"Get metadata for project issue types - Returns the metadata for issue types used for creating issues. Data will not be returned if the user does not have permission to create issues in that project.",
		[]byte(GetCreateIssueMetaProjectIssueTypesInputSchema),
	)
}

// GetCreateIssueMetaProjectIssueTypesHandler is the handler function for the GetCreateIssueMetaProjectIssueTypes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCreateIssueMetaProjectIssueTypesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issue/createmeta/{projectIdOrKey}/issuetypes", args, []string{"projectIdOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetCreateIssueMetaProjectIssueTypes")
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
