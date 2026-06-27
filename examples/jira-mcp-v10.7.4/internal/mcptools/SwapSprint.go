package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SwapSprint tool
const SwapSprintInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The sprint to swap with.\",\n      \"properties\": {\n        \"sprintToSwapWith\": {\n          \"example\": 3,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"swap\": {\n          \"format\": \"int64\",\n          \"type\": \"integer\",\n          \"writeOnly\": true\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"sprintId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// NewSwapSprintMCPTool creates the MCP Tool instance for SwapSprint
func NewSwapSprintMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SwapSprint",
		"Swap the position of two sprints - Swap the position of the sprint with the second sprint.",
		[]byte(SwapSprintInputSchema),
	)
}

// SwapSprintHandler is the handler function for the SwapSprint tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SwapSprintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/sprint/{sprintId}/swap", args, []string{"sprintId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SwapSprint"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
