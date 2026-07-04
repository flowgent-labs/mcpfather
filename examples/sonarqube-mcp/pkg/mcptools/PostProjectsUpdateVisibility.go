package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostProjectsUpdateVisibility tool
const PostProjectsUpdateVisibilityInputSchema = "{\n  \"properties\": {\n    \"project\": {\n      \"description\": \"Project, application or portfolio key\",\n      \"type\": \"string\"\n    },\n    \"visibility\": {\n      \"description\": \"New visibility\",\n      \"enum\": [\n        \"private\",\n        \"public\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"project\",\n    \"visibility\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostProjectsUpdateVisibilityMCPTool creates the MCP Tool instance for PostProjectsUpdateVisibility
func NewPostProjectsUpdateVisibilityMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostProjectsUpdateVisibility",
		"Updates visibility of a project, application or a portfolio. - Updates visibility of a project, application or a portfolio. Requires 'Project administer' permission on the specified entity",
		[]byte(PostProjectsUpdateVisibilityInputSchema),
	)
}

// PostProjectsUpdateVisibilityHandler is the handler function for the PostProjectsUpdateVisibility tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostProjectsUpdateVisibilityHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/projects/update_visibility", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostProjectsUpdateVisibility")
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
