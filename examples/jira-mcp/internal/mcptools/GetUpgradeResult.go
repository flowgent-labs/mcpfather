package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetUpgradeResult tool
const GetUpgradeResultInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetUpgradeResult tool (Status: 200, Content-Type: application/json)
const GetUpgradeResultResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the result of the last upgrade.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **startTime** (Type: string, date-time):\n  - **duration** (Type: integer, int64):\n      - Example: '2001'\n  - **message** (Type: string):\n  - **outcome** (Type: string):\n      - Example: 'SUCCESS'\n"

// NewGetUpgradeResultMCPTool creates the MCP Tool instance for GetUpgradeResult
func NewGetUpgradeResultMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUpgradeResult",
		"Get result of the last upgrade task - Returns the result of the last upgrade task.",
		[]byte(GetUpgradeResultInputSchema),
	)
}

// GetUpgradeResultHandler is the handler function for the GetUpgradeResult tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUpgradeResultHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/upgrade", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetUpgradeResult")
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
