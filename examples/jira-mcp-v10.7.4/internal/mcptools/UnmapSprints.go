package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UnmapSprints tool
const UnmapSprintsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The sprints to unmap.\",\n      \"properties\": {\n        \"sprintIds\": {\n          \"example\": [\n            10001,\n            10004,\n            10005\n          ],\n          \"items\": {\n            \"format\": \"int64\",\n            \"type\": \"integer\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// NewUnmapSprintsMCPTool creates the MCP Tool instance for UnmapSprints
func NewUnmapSprintsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UnmapSprints",
		"Unmap sprints from being synced - Sets the Synced flag to false for all sprints in the provided list.",
		[]byte(UnmapSprintsInputSchema),
	)
}

// UnmapSprintsHandler is the handler function for the UnmapSprints tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UnmapSprintsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/agile/1.0/sprint/unmap", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UnmapSprints"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
