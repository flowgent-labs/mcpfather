package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the MoveIssuesToEpic tool
const MoveIssuesToEpicInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The issues to move to the epic.\",\n      \"properties\": {\n        \"issues\": {\n          \"example\": \"['ISSUE-1', 'ISSUE-2']\",\n          \"items\": {\n            \"example\": \"['ISSUE-1', 'ISSUE-2']\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"epicIdOrKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"epicIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewMoveIssuesToEpicMCPTool creates the MCP Tool instance for MoveIssuesToEpic
func NewMoveIssuesToEpicMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"MoveIssuesToEpic",
		"Move issues to a specific epic - Moves issues to an epic, for a given epic id. Issues can be only in a single epic at the same time. That means that already assigned issues to an epic, will not be assigned to the previous epic anymore. The user needs to have the edit issue permission for all issue they want to move and to the epic. The maximum number of issues that can be moved in one operation is 50.",
		[]byte(MoveIssuesToEpicInputSchema),
	)
}

// MoveIssuesToEpicHandler is the handler function for the MoveIssuesToEpic tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveIssuesToEpicHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/epic/{epicIdOrKey}/issue", args, []string{"epicIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "MoveIssuesToEpic"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
