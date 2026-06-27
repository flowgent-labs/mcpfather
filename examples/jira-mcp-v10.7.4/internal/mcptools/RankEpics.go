package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the RankEpics tool
const RankEpicsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Bean which contains the information where the given epic should be ranked.\",\n      \"properties\": {\n        \"rankAfterEpic\": {\n          \"example\": \"10001\",\n          \"type\": \"string\"\n        },\n        \"rankBeforeEpic\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"rankCustomFieldId\": {\n          \"example\": 10521,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"epicIdOrKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"epicIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewRankEpicsMCPTool creates the MCP Tool instance for RankEpics
func NewRankEpicsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RankEpics",
		"Rank an epic relative to another - Moves (ranks) an epic before or after a given epic. If rankCustomFieldId is not defined, the default rank field will be used.",
		[]byte(RankEpicsInputSchema),
	)
}

// RankEpicsHandler is the handler function for the RankEpics tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RankEpicsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/agile/1.0/epic/{epicIdOrKey}/rank", args, []string{"epicIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RankEpics"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
