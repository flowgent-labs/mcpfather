package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostApplicationsDeleteBranch tool
const PostApplicationsDeleteBranchInputSchema = "{\n  \"properties\": {\n    \"application\": {\n      \"description\": \"Application key\",\n      \"type\": \"string\"\n    },\n    \"branch\": {\n      \"description\": \"Branch name\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"application\",\n    \"branch\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostApplicationsDeleteBranchMCPTool creates the MCP Tool instance for PostApplicationsDeleteBranch
func NewPostApplicationsDeleteBranchMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostApplicationsDeleteBranch",
		"Delete a branch on a given application. - Delete a branch on a given application. Requires 'Administrator' permission on the application",
		[]byte(PostApplicationsDeleteBranchInputSchema),
	)
}

// PostApplicationsDeleteBranchHandler is the handler function for the PostApplicationsDeleteBranch tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostApplicationsDeleteBranchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/applications/delete_branch", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostApplicationsDeleteBranch")
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
