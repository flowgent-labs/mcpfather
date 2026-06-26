package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the PartiallyUpdateSprint tool
const PartiallyUpdateSprintInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The updated sprint.\",\n      \"properties\": {\n        \"activatedDate\": {\n          \"example\": \"2015-04-11T15:22:00.000+10:00\",\n          \"type\": \"string\"\n        },\n        \"autoStartStop\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"completeDate\": {\n          \"example\": \"2015-04-20T11:04:00.000+10:00\",\n          \"type\": \"string\"\n        },\n        \"endDate\": {\n          \"example\": \"2015-04-20T01:22:00.000+10:00\",\n          \"type\": \"string\"\n        },\n        \"goal\": {\n          \"example\": \"Goal for the sprint\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": 10001,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"incompleteIssuesDestinationId\": {\n          \"example\": 10001,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"name\": {\n          \"example\": \"Sprint 1\",\n          \"type\": \"string\"\n        },\n        \"originBoardId\": {\n          \"example\": 5,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/agile/1.0/sprint/10001\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"startDate\": {\n          \"example\": \"2015-04-11T15:22:00.000+10:00\",\n          \"type\": \"string\"\n        },\n        \"state\": {\n          \"example\": \"active\",\n          \"type\": \"string\"\n        },\n        \"synced\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"sprintId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PartiallyUpdateSprint tool (Status: 200, Content-Type: application/json)
const PartiallyUpdateSprintResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the updated sprint.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **startDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **activatedDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **completeDate** (Type: string):\n      - Example: '2015-04-20T11:04:00.000+10:00'\n  - **originBoardId** (Type: integer, int64):\n      - Example: '5'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/sprint/10001'\n  - **goal** (Type: string):\n      - Example: 'Goal for the sprint'\n  - **state** (Type: string):\n      - Example: 'active'\n  - **endDate** (Type: string):\n      - Example: '2015-04-20T01:22:00.000+10:00'\n  - **synced** (Type: boolean):\n      - Example: 'true'\n  - **autoStartStop** (Type: boolean):\n      - Example: 'true'\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **incompleteIssuesDestinationId** (Type: integer, int64):\n      - Example: '10001'\n  - **name** (Type: string):\n      - Example: 'Sprint 1'\n"

// NewPartiallyUpdateSprintMCPTool creates the MCP Tool instance for PartiallyUpdateSprint
func NewPartiallyUpdateSprintMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PartiallyUpdateSprint",
		"Partially update a sprint - Performs a partial update of a sprint.\nA partial update means that fields not present in the request JSON will not be updated.\nNotes:\n- Sprints that are in a closed state cannot be updated.\n- A sprint can be started by updating the state to 'active'. This requires the sprint to be in the 'future' state and have a startDate and endDate set.\n- A sprint can be completed by updating the state to 'closed'. This action requires the sprint to be in the 'active' state. This sets the completeDate to the time of the request.\n  If the sprint has offending issues (those which are complete, but have incomplete subtasks) it cannot be closed.\n  If issues are moved to new sprint user has to have issues edit permissions.\n- Other changes to state are not allowed.\n- The completeDate field cannot be updated manually.\n- Sprint goal can be removed by updating it's value to empty string\n- Only Jira administrators can edit dates on sprints that are marked as synced.",
		[]byte(PartiallyUpdateSprintInputSchema),
	)
}

// PartiallyUpdateSprintHandler is the handler function for the PartiallyUpdateSprint tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PartiallyUpdateSprintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/sprint/{sprintId}", args, []string{"sprintId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "PartiallyUpdateSprint"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
