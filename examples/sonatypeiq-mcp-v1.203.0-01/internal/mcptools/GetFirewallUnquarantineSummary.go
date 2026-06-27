package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetFirewallUnquarantineSummary tool
const GetFirewallUnquarantineSummaryInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetFirewallUnquarantineSummary tool (Status: 200, Content-Type: application/json)
const GetFirewallUnquarantineSummaryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "autoReleaseQuarantineCountMTD" + "\x60" + " is the number of auto-released quarantine components from the start of the current month to the current date.</li><li>" + "\x60" + "autoReleaseQuarantineCountYTD" + "\x60" + " is the number of auto-released quarantine components from the start of the current year to the current date.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **autoReleaseQuarantineCountYTD** (Type: integer, int64):\n  - **autoReleaseQuarantineCountMTD** (Type: integer, int64):\n"

// NewGetFirewallUnquarantineSummaryMCPTool creates the MCP Tool instance for GetFirewallUnquarantineSummary
func NewGetFirewallUnquarantineSummaryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetFirewallUnquarantineSummary",
		"Use this method to track how many components have been automatically released from quarantine over different time periods.\n\nPermissions required: View IQ Elements",
		[]byte(GetFirewallUnquarantineSummaryInputSchema),
	)
}

// GetFirewallUnquarantineSummaryHandler is the handler function for the GetFirewallUnquarantineSummary tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetFirewallUnquarantineSummaryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/releaseQuarantine/summary", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetFirewallUnquarantineSummary")
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
