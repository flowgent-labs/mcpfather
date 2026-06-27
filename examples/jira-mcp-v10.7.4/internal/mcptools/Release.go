package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the Release tool
const ReleaseInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"No request body is needed\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewReleaseMCPTool creates the MCP Tool instance for Release
func NewReleaseMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Release",
		"Invalidate the current WebSudo session - This method invalidates the any current WebSudo session.",
		[]byte(ReleaseInputSchema),
	)
}

// ReleaseHandler is the handler function for the Release tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ReleaseHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/auth/1/websudo", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Release"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
