package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssuesForSprint tool
const GetIssuesForSprintInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"expand\": {\n      \"type\": \"string\"\n    },\n    \"fields\": {\n      \"items\": {\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    },\n    \"jql\": {\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"sprintId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"startAt\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"validateQuery\": {\n      \"type\": \"boolean\"\n    }\n  },\n  \"required\": [\n    \"boardId\",\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssuesForSprint tool (Status: 200, Content-Type: application/json)
const GetIssuesForSprintResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested issues, at the specified page of the results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **activatedDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **completeDate** (Type: string):\n      - Example: '2015-04-20T11:04:00.000+10:00'\n  - **originBoardId** (Type: integer, int64):\n      - Example: '5'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/sprint/10001'\n  - **goal** (Type: string):\n      - Example: 'Goal for the sprint'\n  - **state** (Type: string):\n      - Example: 'active'\n  - **endDate** (Type: string):\n      - Example: '2015-04-20T01:22:00.000+10:00'\n  - **synced** (Type: boolean):\n      - Example: 'true'\n  - **autoStartStop** (Type: boolean):\n      - Example: 'true'\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **incompleteIssuesDestinationId** (Type: integer, int64):\n      - Example: '10001'\n  - **name** (Type: string):\n      - Example: 'Sprint 1'\n  - **startDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n"

// NewGetIssuesForSprintMCPTool creates the MCP Tool instance for GetIssuesForSprint
func NewGetIssuesForSprintMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssuesForSprint",
		"Get all issues for a sprint - Get all issues you have access to that belong to the sprint from the board. Issue returned from this resource contains additional fields like: sprint, closedSprints, flagged and epic. Issues are returned ordered by rank. JQL order has higher priority than default rank.",
		[]byte(GetIssuesForSprintInputSchema),
	)
}

// GetIssuesForSprintHandler is the handler function for the GetIssuesForSprint tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssuesForSprintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/sprint/{sprintId}/issue", args, []string{"boardId", "sprintId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssuesForSprint"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
