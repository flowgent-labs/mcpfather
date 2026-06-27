package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the ListIndexSnapshot tool
const ListIndexSnapshotInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the ListIndexSnapshot tool (Status: 200, Content-Type: application/json)
const ListIndexSnapshotResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the list consisting of absolute paths to currently available index snapshots\n\n## Response Structure\n\n- Structure (Type: object):\n  - **absolutePath** (Type: string):\n      - Example: '/var/atlassian/application-data/jira/caches/indexesV1/issue'\n  - **timestamp** (Type: integer, int64):\n      - Example: '1612345678900'\n"

// NewListIndexSnapshotMCPTool creates the MCP Tool instance for ListIndexSnapshot
func NewListIndexSnapshotMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ListIndexSnapshot",
		"Get list of available index snapshots - Lists available index snapshots absolute paths with timestamps",
		[]byte(ListIndexSnapshotInputSchema),
	)
}

// ListIndexSnapshotHandler is the handler function for the ListIndexSnapshot tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ListIndexSnapshotHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/index-snapshot", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ListIndexSnapshot"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
