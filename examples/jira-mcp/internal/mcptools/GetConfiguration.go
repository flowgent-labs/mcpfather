package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration tool
const GetConfigurationInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"boardId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration tool (Status: 200, Content-Type: application/json)
const GetConfigurationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the configuration of the board for given boardId.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'Scrum'\n  - **ranking** (Type: object):\n    - **rankCustomFieldId** (Type: integer, int64):\n        - Example: '10020'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/board/10001'\n  - **filter** (Type: object):\n    - **self** (Type: string, uri):\n        - Example: 'http://www.example.com/jira/rest/agile/1.0/filter/1001'\n    - **id** (Type: string):\n        - Example: '1001'\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **columnConfig** (Type: object):\n    - **constraintType** (Type: string):\n        - Example: 'issueCount'\n    - **columns** (Type: array):\n      - **Items** (Type: object):\n        - **max** (Type: integer, int32):\n            - Example: '4'\n        - **min** (Type: integer, int32):\n            - Example: '2'\n        - **name** (Type: string):\n            - Example: 'To Do'\n        - **statuses** (Type: array):\n          - **[cyclic reference]**\n  - **estimation** (Type: object):\n    - **field** (Type: object):\n      - **displayName** (Type: string):\n          - Example: 'Story Points'\n      - **fieldId** (Type: string):\n          - Example: 'customfield_10002'\n    - **type** (Type: string):\n        - Example: 'field'\n  - **type** (Type: string):\n      - Example: 'scrum'\n  - **subQuery** (Type: object):\n    - **query** (Type: string):\n        - Example: 'project = HSP'\n"

// NewGetConfigurationMCPTool creates the MCP Tool instance for GetConfiguration
func NewGetConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration",
		"Get the board configuration - Get the board configuration.\nThe response contains the following fields:\n- id - Id of the board.\n- name - Name of the board.\n- filter - Reference to the filter used by the given board.\n- subQuery (Kanban only) - JQL subquery used by the given board.\n- columnConfig - The column configuration lists the columns for the board, in the order defined in the column configuration.\nFor each column, it shows the issue status mapping\nas well as the constraint type (Valid values: none, issueCount, issueCountExclSubs) for the min/max number of issues.\nNote, the last column with statuses mapped to it is treated as the \"Done\" column,\nwhich means that issues in that column will be marked as already completed.\n- estimation (Scrum only) - Contains information about type of estimation used for the board. Valid values: none, issueCount, field.\nIf the estimation type is \"field\", the Id and display name of the field used for estimation is also returned.\nNote, estimates for an issue can be updated by a PUT /rest/api/2/issue/{issueIdOrKey} request, however the fields must be on the screen.\n\"timeoriginalestimate\" field will never be on the screen, so in order to update it \"originalEstimate\" in \"timetracking\" field should be updated.\n- ranking - Contains information about custom field used for ranking in the given board.",
		[]byte(GetConfigurationInputSchema),
	)
}

// GetConfigurationHandler is the handler function for the GetConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/configuration", args, []string{"boardId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration")
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
