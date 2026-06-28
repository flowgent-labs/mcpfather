package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteConfiguration1 tool
const DeleteConfiguration1InputSchema = "{\n  \"properties\": {\n    \"property\": {\n      \"description\": \"Enter the names of the system properties. Values provided for name are case-sensitive.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"type\": \"object\"\n}"

// NewDeleteConfiguration1MCPTool creates the MCP Tool instance for DeleteConfiguration1
func NewDeleteConfiguration1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteConfiguration1",
		"Use this method to disable one or more IQ Server system properties. The property names are case-sensitive.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(DeleteConfiguration1InputSchema),
	)
}

// DeleteConfiguration1Handler is the handler function for the DeleteConfiguration1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteConfiguration1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/config", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteConfiguration1")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
