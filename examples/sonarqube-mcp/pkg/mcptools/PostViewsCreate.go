package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostViewsCreate tool
const PostViewsCreateInputSchema = "{\n  \"properties\": {\n    \"description\": {\n      \"description\": \"Description for the new portfolio, can be left blank\",\n      \"maxLength\": 256,\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"Key for the new portfolio. A suitable key will be generated if not provided\",\n      \"maxLength\": 400,\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"Name for the new portfolio\",\n      \"maxLength\": 256,\n      \"type\": \"string\"\n    },\n    \"parent\": {\n      \"description\": \"Key of the parent portfolio, when creating a sub portfolio\",\n      \"maxLength\": 400,\n      \"type\": \"string\"\n    },\n    \"visibility\": {\n      \"description\": \"Whether the created portfolio or application should be visible to everyone, or only specific user/groups. Only applies to root portfolios. If no visibility is specified, the default visibility will be used.\",\n      \"enum\": [\n        \"private\",\n        \"public\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"name\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostViewsCreateMCPTool creates the MCP Tool instance for PostViewsCreate
func NewPostViewsCreateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostViewsCreate",
		"Create a new portfolio. - Create a new portfolio. Requires 'Administer System' permission or 'Create Portfolios' permission,",
		[]byte(PostViewsCreateInputSchema),
	)
}

// PostViewsCreateHandler is the handler function for the PostViewsCreate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostViewsCreateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/views/create", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostViewsCreate")
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
