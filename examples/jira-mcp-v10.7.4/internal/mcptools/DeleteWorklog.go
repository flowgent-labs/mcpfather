package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteWorklog tool
const DeleteWorklogInputSchema = "{\n  \"properties\": {\n    \"adjustEstimate\": {\n      \"description\": \"Allows you to provide specific instructions to update the remaining time estimate of the issue. Valid values are: new, leave, manual, auto\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"Id of the worklog to be deleted\",\n      \"type\": \"string\"\n    },\n    \"increaseBy\": {\n      \"description\": \"Required when 'manual' is selected for adjustEstimate. e.g. \\\"2d\\\"\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"a string containing the issue id or key the worklog belongs to\",\n      \"type\": \"string\"\n    },\n    \"newEstimate\": {\n      \"description\": \"Required when 'new' is selected for adjustEstimate. e.g. \\\"2d\\\"\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteWorklogMCPTool creates the MCP Tool instance for DeleteWorklog
func NewDeleteWorklogMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteWorklog",
		"Delete a worklog entry - Deletes an existing worklog entry.",
		[]byte(DeleteWorklogInputSchema),
	)
}

// DeleteWorklogHandler is the handler function for the DeleteWorklog tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteWorklogHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issue/{issueIdOrKey}/worklog/{id}", args, []string{"id", "issueIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteWorklog"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
