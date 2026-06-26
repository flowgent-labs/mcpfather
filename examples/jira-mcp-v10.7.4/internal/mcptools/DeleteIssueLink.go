package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteIssueLink tool
const DeleteIssueLinkInputSchema = "{\n  \"properties\": {\n    \"linkId\": {\n      \"description\": \"The issue link id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"linkId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteIssueLinkMCPTool creates the MCP Tool instance for DeleteIssueLink
func NewDeleteIssueLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteIssueLink",
		"Delete an issue link with the specified id - Deletes an issue link with the specified id.",
		[]byte(DeleteIssueLinkInputSchema),
	)
}

// DeleteIssueLinkHandler is the handler function for the DeleteIssueLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteIssueLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issueLink/{linkId}", args, []string{"linkId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteIssueLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
