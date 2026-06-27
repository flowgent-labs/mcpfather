package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetPinComment tool
const SetPinCommentInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"'true' must be included as raw data\",\n      \"type\": \"boolean\"\n    },\n    \"id\": {\n      \"description\": \"Comment id\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"id\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetPinCommentMCPTool creates the MCP Tool instance for SetPinComment
func NewSetPinCommentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetPinComment",
		"Pin a comment - Pins a comment to the top of the comment list.",
		[]byte(SetPinCommentInputSchema),
	)
}

// SetPinCommentHandler is the handler function for the SetPinComment tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetPinCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issue/{issueIdOrKey}/comment/{id}/pin", args, []string{"id", "issueIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetPinComment"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
