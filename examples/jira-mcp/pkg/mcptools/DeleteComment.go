package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/pkg/helpers"
	"time"
)

// Input Schema for the DeleteComment tool
const DeleteCommentInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"Comment id\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteCommentMCPTool creates the MCP Tool instance for DeleteComment
func NewDeleteCommentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteComment",
		"Delete a comment - Deletes an existing comment.",
		[]byte(DeleteCommentInputSchema),
	)
}

// DeleteCommentHandler is the handler function for the DeleteComment tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issue/{issueIdOrKey}/comment/{id}", args, []string{"id", "issueIdOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteComment")
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
