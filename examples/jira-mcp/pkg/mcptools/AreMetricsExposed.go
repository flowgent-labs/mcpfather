package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/pkg/helpers"
	"time"
)

// Input Schema for the AreMetricsExposed tool
const AreMetricsExposedInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the AreMetricsExposed tool (Status: 200, Content-Type: application/json)
const AreMetricsExposedResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns whether JMX metrics are being exposed.\n\n## Response Structure\n\n- Structure (Type: boolean):\n"

// NewAreMetricsExposedMCPTool creates the MCP Tool instance for AreMetricsExposed
func NewAreMetricsExposedMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AreMetricsExposed",
		"Check if JMX metrics are being exposed - Checks if JMX metrics are being exposed",
		[]byte(AreMetricsExposedInputSchema),
	)
}

// AreMetricsExposedHandler is the handler function for the AreMetricsExposed tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AreMetricsExposedHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/monitoring/jmx/areMetricsExposed", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AreMetricsExposed")
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
