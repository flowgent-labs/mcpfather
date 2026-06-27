package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteSprint tool
const DeleteSprintInputSchema = "{\n  \"properties\": {\n    \"sprintId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteSprintMCPTool creates the MCP Tool instance for DeleteSprint
func NewDeleteSprintMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteSprint",
		"Delete a sprint - Deletes a sprint. Once a sprint is deleted, all issues in the sprint will be moved to the backlog. To delete a synced sprint, you must unsync it first.",
		[]byte(DeleteSprintInputSchema),
	)
}

// DeleteSprintHandler is the handler function for the DeleteSprint tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteSprintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/agile/1.0/sprint/{sprintId}", args, []string{"sprintId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteSprint"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
