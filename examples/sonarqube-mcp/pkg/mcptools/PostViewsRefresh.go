package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostViewsRefresh tool
const PostViewsRefreshInputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"Key: Root portfolio key. If not specified, all portfolios are refreshed.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewPostViewsRefreshMCPTool creates the MCP Tool instance for PostViewsRefresh
func NewPostViewsRefreshMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostViewsRefresh",
		"Refresh one or all of the portfolios. - Refresh one or all of the portfolios. When a key is not specified, all portfolios are refreshed. Requires one of the following permissions: * 'Administer System' when key is not specified * 'Administer' rights on the specified portfolio when key is specified",
		[]byte(PostViewsRefreshInputSchema),
	)
}

// PostViewsRefreshHandler is the handler function for the PostViewsRefresh tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostViewsRefreshHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/views/refresh", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostViewsRefresh")
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
