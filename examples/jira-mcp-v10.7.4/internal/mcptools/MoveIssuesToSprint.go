package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the MoveIssuesToSprint tool
const MoveIssuesToSprintInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The issues to move.\",\n      \"properties\": {\n        \"issues\": {\n          \"example\": \"['ISSUE-1', 'ISSUE-2']\",\n          \"items\": {\n            \"example\": \"['ISSUE-1', 'ISSUE-2']\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"sprintId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// NewMoveIssuesToSprintMCPTool creates the MCP Tool instance for MoveIssuesToSprint
func NewMoveIssuesToSprintMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"MoveIssuesToSprint",
		"Move issues to a sprint - Moves issues to a sprint, for a given sprint Id. Issues can only be moved to open or active sprints. The maximum number of issues that can be moved in one operation is 50.",
		[]byte(MoveIssuesToSprintInputSchema),
	)
}

// MoveIssuesToSprintHandler is the handler function for the MoveIssuesToSprint tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveIssuesToSprintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/sprint/{sprintId}/issue", args, []string{"sprintId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "MoveIssuesToSprint"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
