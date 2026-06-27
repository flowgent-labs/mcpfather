package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the MoveIssuesToBacklog tool
const MoveIssuesToBacklogInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The issues to move to the backlog.\",\n      \"properties\": {\n        \"issues\": {\n          \"example\": \"['ISSUE-1', 'ISSUE-2']\",\n          \"items\": {\n            \"example\": \"['ISSUE-1', 'ISSUE-2']\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// NewMoveIssuesToBacklogMCPTool creates the MCP Tool instance for MoveIssuesToBacklog
func NewMoveIssuesToBacklogMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"MoveIssuesToBacklog",
		"Update issues to move them to the backlog - Move issues to the backlog. This operation is equivalent to remove future and active sprints from a given set of issues. At most 50 issues may be moved at once.",
		[]byte(MoveIssuesToBacklogInputSchema),
	)
}

// MoveIssuesToBacklogHandler is the handler function for the MoveIssuesToBacklog tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveIssuesToBacklogHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/backlog/issue", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "MoveIssuesToBacklog"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
