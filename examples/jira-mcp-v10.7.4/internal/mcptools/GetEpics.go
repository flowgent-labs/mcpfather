package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetEpics tool
const GetEpicsInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"done\": {\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"startAt\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"boardId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetEpics tool (Status: 200, Content-Type: application/json)
const GetEpicsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested epics, at the specified page of the results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10000'\n  - **summary** (Type: string):\n      - Example: 'Epic 1 summary'\n  - **color** (Type: object):\n      - Example: '\"color_6\"'\n    - **key** (Type: string):\n        - Example: 'ghx-label-1'\n        - Enum: ['color_1', 'color_2', 'color_3', 'color_4', 'color_5', 'color_6', 'color_7', 'color_8', 'color_9', 'color_10', 'color_11', 'color_12', 'color_13', 'color_14']\n  - **done** (Type: boolean):\n      - Example: 'true'\n  - **id** (Type: integer, int64):\n      - Example: '10000'\n  - **key** (Type: string):\n      - Example: 'PR-1'\n  - **name** (Type: string):\n      - Example: 'Epic 1'\n"

// NewGetEpicsMCPTool creates the MCP Tool instance for GetEpics
func NewGetEpicsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetEpics",
		"Get all epics from the board - Returns all epics from the board, for the given board Id. This only includes epics that the user has permission to view. Note, if the user does not have permission to view the board, no epics will be returned at all.",
		[]byte(GetEpicsInputSchema),
	)
}

// GetEpicsHandler is the handler function for the GetEpics tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetEpicsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/epic", args, []string{"boardId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetEpics"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
