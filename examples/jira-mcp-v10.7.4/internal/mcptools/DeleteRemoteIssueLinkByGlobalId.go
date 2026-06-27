package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteRemoteIssueLinkByGlobalId tool
const DeleteRemoteIssueLinkByGlobalIdInputSchema = "{\n  \"properties\": {\n    \"globalId\": {\n      \"description\": \"Global id of the remote issue link\",\n      \"type\": \"string\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"globalId\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteRemoteIssueLinkByGlobalIdMCPTool creates the MCP Tool instance for DeleteRemoteIssueLinkByGlobalId
func NewDeleteRemoteIssueLinkByGlobalIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteRemoteIssueLinkByGlobalId",
		"Delete remote issue link - Delete the remote issue link with the given global id on the issue.",
		[]byte(DeleteRemoteIssueLinkByGlobalIdInputSchema),
	)
}

// DeleteRemoteIssueLinkByGlobalIdHandler is the handler function for the DeleteRemoteIssueLinkByGlobalId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteRemoteIssueLinkByGlobalIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issue/{issueIdOrKey}/remotelink", args, []string{"issueIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteRemoteIssueLinkByGlobalId"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
