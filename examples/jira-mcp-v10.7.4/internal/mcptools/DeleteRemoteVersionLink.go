package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteRemoteVersionLink tool
const DeleteRemoteVersionLinkInputSchema = "{\n  \"properties\": {\n    \"globalId\": {\n      \"description\": \"The id of the remote issue link to be deleted.\",\n      \"type\": \"string\"\n    },\n    \"versionId\": {\n      \"description\": \"ID of the version.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"globalId\",\n    \"versionId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteRemoteVersionLinkMCPTool creates the MCP Tool instance for DeleteRemoteVersionLink
func NewDeleteRemoteVersionLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteRemoteVersionLink",
		"Delete specific remote version link - Delete a specific remote version link with the given version ID and global ID.",
		[]byte(DeleteRemoteVersionLinkInputSchema),
	)
}

// DeleteRemoteVersionLinkHandler is the handler function for the DeleteRemoteVersionLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteRemoteVersionLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/version/{versionId}/remotelink/{globalId}", args, []string{"globalId", "versionId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteRemoteVersionLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
