package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration1 tool
const GetConfiguration1InputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration1 tool (Status: 200, Content-Type: application/json)
const GetConfiguration1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned the configuration of optional features in Jira.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **timeTrackingEnabled** (Type: boolean):\n      - Example: 'true'\n  - **unassignedIssuesAllowed** (Type: boolean):\n      - Example: 'false'\n  - **votingEnabled** (Type: boolean):\n      - Example: 'true'\n  - **watchingEnabled** (Type: boolean):\n      - Example: 'true'\n  - **attachmentsEnabled** (Type: boolean):\n      - Example: 'true'\n  - **issueLinkingEnabled** (Type: boolean):\n      - Example: 'true'\n  - **subTasksEnabled** (Type: boolean):\n      - Example: 'false'\n  - **timeTrackingConfiguration** (Type: object):\n    - **timeFormat** (Type: string):\n        - Example: 'pretty'\n        - Enum: ['pretty', 'days', 'hours']\n    - **workingDaysPerWeek** (Type: number, double):\n        - Example: '5'\n    - **workingHoursPerDay** (Type: number, double):\n        - Example: '8'\n    - **defaultUnit** (Type: string):\n        - Example: 'day'\n        - Enum: ['minute', 'hour', 'day', 'week']\n"

// NewGetConfiguration1MCPTool creates the MCP Tool instance for GetConfiguration1
func NewGetConfiguration1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration1",
		"Get Jira configuration details - Returns the information if the optional features in Jira are enabled or disabled. If the time tracking is enabled, it also returns the detailed information about time tracking configuration.",
		[]byte(GetConfiguration1InputSchema),
	)
}

// GetConfiguration1Handler is the handler function for the GetConfiguration1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguration1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/configuration", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration1")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
