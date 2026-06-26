package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the CreateIndexSnapshot tool
const CreateIndexSnapshotInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the CreateIndexSnapshot tool (Status: 202, Content-Type: application/json)
const CreateIndexSnapshotResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 202\n\n**Content-Type:** application/json\n\n> Returns the absolute path which index snapshot will be placed in, after it's created\n\n## Response Structure\n\n- Structure (Type: object):\n  - **futureAbsolutePath** (Type: string):\n      - Example: '/home/atlassian/shared-home/export/indexsnapshots/IndexSnapshot_2021-Jul-21--2142-34-601.tar.sz'\n"

// NewCreateIndexSnapshotMCPTool creates the MCP Tool instance for CreateIndexSnapshot
func NewCreateIndexSnapshotMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateIndexSnapshot",
		"Create index snapshot if not in progress - Starts taking an index snapshot if no other snapshot creation process is in progress",
		[]byte(CreateIndexSnapshotInputSchema),
	)
}

// CreateIndexSnapshotHandler is the handler function for the CreateIndexSnapshot tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateIndexSnapshotHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/index-snapshot", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateIndexSnapshot"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
