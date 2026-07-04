package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostUsersSetAiToolUsage tool
const PostUsersSetAiToolUsageInputSchema = "{\n  \"properties\": {\n    \"project\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"project\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostUsersSetAiToolUsageMCPTool creates the MCP Tool instance for PostUsersSetAiToolUsage
func NewPostUsersSetAiToolUsageMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostUsersSetAiToolUsage",
		"Indicates whether the current user utilizes any tool that generates AI code for a given project. - Indicates whether the current user utilizes any tool that generates AI code for a given project. Requires 'Browse' permission on the specified project.",
		[]byte(PostUsersSetAiToolUsageInputSchema),
	)
}

// PostUsersSetAiToolUsageHandler is the handler function for the PostUsersSetAiToolUsage tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostUsersSetAiToolUsageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/users/set_ai_tool_usage", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostUsersSetAiToolUsage")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
