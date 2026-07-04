package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostViewsUpdate tool
const PostViewsUpdateInputSchema = "{\n  \"properties\": {\n    \"description\": {\n      \"description\": \"New description for the portfolio\",\n      \"maxLength\": 256,\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"Key of the portfolio to update\",\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"New name for the portfolio\",\n      \"maxLength\": 256,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"name\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostViewsUpdateMCPTool creates the MCP Tool instance for PostViewsUpdate
func NewPostViewsUpdateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostViewsUpdate",
		"Update a portfolio. - Update a portfolio. Requires 'Administrator' permission on the portfolio.",
		[]byte(PostViewsUpdateInputSchema),
	)
}

// PostViewsUpdateHandler is the handler function for the PostViewsUpdate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostViewsUpdateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/views/update", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostViewsUpdate")
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
