package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAllSprints tool
const GetAllSprintsInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"description\": \"The Id of the board that contains the requested sprints.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of sprints to return per page. Default: 50.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"startAt\": {\n      \"description\": \"The starting index of the returned sprints. Base index: 0.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"state\": {\n      \"description\": \"Filters results to sprints in specified states. Valid values: future, active, closed. You can define multiple states separated by commas, e.g. state=active,closed\",\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"boardId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAllSprints tool (Status: 200, Content-Type: application/json)
const GetAllSprintsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested sprints, at the specified page of the results. Sprints will be ordered first by state (i.e. closed, active, future) then by their position in the backlog.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **autoStartStop** (Type: boolean):\n      - Example: 'true'\n  - **endDate** (Type: string):\n      - Example: '2015-04-20T01:22:00.000+10:00'\n  - **name** (Type: string):\n      - Example: 'Sprint 1'\n  - **completeDate** (Type: string):\n      - Example: '2015-04-20T11:04:00.000+10:00'\n  - **goal** (Type: string):\n      - Example: 'Goal for the sprint'\n  - **incompleteIssuesDestinationId** (Type: integer, int64):\n      - Example: '10001'\n  - **originBoardId** (Type: integer, int64):\n      - Example: '5'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/sprint/10001'\n  - **startDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **activatedDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **state** (Type: string):\n      - Example: 'active'\n  - **synced** (Type: boolean):\n      - Example: 'true'\n"

// NewGetAllSprintsMCPTool creates the MCP Tool instance for GetAllSprints
func NewGetAllSprintsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllSprints",
		"Get all sprints from a board - Returns all sprints from a board, for a given board Id. This only includes sprints that the user has permission to view.",
		[]byte(GetAllSprintsInputSchema),
	)
}

// GetAllSprintsHandler is the handler function for the GetAllSprints tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllSprintsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/sprint", args, []string{"boardId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAllSprints")
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
