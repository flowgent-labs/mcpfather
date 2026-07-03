package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostIssuesReindex tool
const PostIssuesReindexInputSchema = "{\n  \"properties\": {\n    \"project\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"project\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostIssuesReindexMCPTool creates the MCP Tool instance for PostIssuesReindex
func NewPostIssuesReindexMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostIssuesReindex",
		"Reindex issues for a project. - Reindex issues for a project. Require 'Administer System' permission.",
		[]byte(PostIssuesReindexInputSchema),
	)
}

// PostIssuesReindexHandler is the handler function for the PostIssuesReindex tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostIssuesReindexHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/issues/reindex", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostIssuesReindex")
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
