package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the MoveSubTasks tool
const MoveSubTasksInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The description of previous and current position of subtask in the sequence.\",\n      \"properties\": {\n        \"current\": {\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"original\": {\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"The parent issue's key or id\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewMoveSubTasksMCPTool creates the MCP Tool instance for MoveSubTasks
func NewMoveSubTasksMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"MoveSubTasks",
		"Reorder an issue's subtasks - Reorders an issue's subtasks by moving the subtask at index 'from' to index 'to'.",
		[]byte(MoveSubTasksInputSchema),
	)
}

// MoveSubTasksHandler is the handler function for the MoveSubTasks tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveSubTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issue/{issueIdOrKey}/subtask/move", args, []string{"issueIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "MoveSubTasks"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
