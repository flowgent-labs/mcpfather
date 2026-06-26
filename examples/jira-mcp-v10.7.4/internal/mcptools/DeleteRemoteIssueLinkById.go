package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteRemoteIssueLinkById tool
const DeleteRemoteIssueLinkByIdInputSchema = "{\n  \"properties\": {\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    },\n    \"linkId\": {\n      \"description\": \"Id of the remote issue link\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\",\n    \"linkId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteRemoteIssueLinkByIdMCPTool creates the MCP Tool instance for DeleteRemoteIssueLinkById
func NewDeleteRemoteIssueLinkByIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteRemoteIssueLinkById",
		"Delete remote issue link by id - Delete the remote issue link with the given id on the issue.",
		[]byte(DeleteRemoteIssueLinkByIdInputSchema),
	)
}

// DeleteRemoteIssueLinkByIdHandler is the handler function for the DeleteRemoteIssueLinkById tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteRemoteIssueLinkByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issue/{issueIdOrKey}/remotelink/{linkId}", args, []string{"issueIdOrKey", "linkId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteRemoteIssueLinkById"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
