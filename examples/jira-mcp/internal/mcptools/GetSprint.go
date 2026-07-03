package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetSprint tool
const GetSprintInputSchema = "{\n  \"properties\": {\n    \"sprintId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetSprint tool (Status: 200, Content-Type: application/json)
const GetSprintResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested sprint.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **startDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **activatedDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **completeDate** (Type: string):\n      - Example: '2015-04-20T11:04:00.000+10:00'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/sprint/10001'\n  - **autoStartStop** (Type: boolean):\n      - Example: 'true'\n  - **goal** (Type: string):\n      - Example: 'Goal for the sprint'\n  - **synced** (Type: boolean):\n      - Example: 'true'\n  - **endDate** (Type: string):\n      - Example: '2015-04-20T01:22:00.000+10:00'\n  - **state** (Type: string):\n      - Example: 'active'\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **incompleteIssuesDestinationId** (Type: integer, int64):\n      - Example: '10001'\n  - **name** (Type: string):\n      - Example: 'Sprint 1'\n  - **originBoardId** (Type: integer, int64):\n      - Example: '5'\n"

// NewGetSprintMCPTool creates the MCP Tool instance for GetSprint
func NewGetSprintMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSprint",
		"Get sprint by id - Returns a single sprint, for a given sprint Id. The sprint will only be returned if the user can view the board that the sprint was created on, or view at least one of the issues in the sprint.",
		[]byte(GetSprintInputSchema),
	)
}

// GetSprintHandler is the handler function for the GetSprint tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSprintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/sprint/{sprintId}", args, []string{"sprintId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSprint")
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
