package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the IsIndexSnapshotRunning tool
const IsIndexSnapshotRunningInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the IsIndexSnapshotRunning tool (Status: 200, Content-Type: application/json)
const IsIndexSnapshotRunningResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns status of current snapshot creation\n\n## Response Structure\n\n- Structure (Type: object):\n  - **running** (Type: boolean):\n      - Example: 'true'\n"

// NewIsIndexSnapshotRunningMCPTool creates the MCP Tool instance for IsIndexSnapshotRunning
func NewIsIndexSnapshotRunningMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"IsIndexSnapshotRunning",
		"Get index snapshot creation status - Checks if index snapshot creation is currently running",
		[]byte(IsIndexSnapshotRunningInputSchema),
	)
}

// IsIndexSnapshotRunningHandler is the handler function for the IsIndexSnapshotRunning tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func IsIndexSnapshotRunningHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/index-snapshot/isRunning", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "IsIndexSnapshotRunning"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
