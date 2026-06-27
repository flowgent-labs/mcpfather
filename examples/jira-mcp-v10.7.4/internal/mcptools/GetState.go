package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetState tool
const GetStateInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetState tool (Status: 200, Content-Type: application/json)
const GetStateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the current state of the cluster upgrade.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **state** (Type: string):\n      - Example: 'UPGRADE_IN_PROGRESS'\n      - Enum: ['STABLE', 'READY_TO_UPGRADE', 'MIXED', 'READY_TO_RUN_UPGRADE_TASKS', 'RUNNING_UPGRADE_TASKS', 'UPGRADE_TASKS_FAILED']\n  - **build** (Type: object):\n    - **version** (Type: string):\n        - Example: '8.0.0'\n    - **buildNumber** (Type: integer, int64):\n        - Example: '1000'\n"

// NewGetStateMCPTool creates the MCP Tool instance for GetState
func NewGetStateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetState",
		"Get cluster upgrade state - Returns the current state of the cluster upgrade.",
		[]byte(GetStateInputSchema),
	)
}

// GetStateHandler is the handler function for the GetState tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetStateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/cluster/zdu/state", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetState"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
