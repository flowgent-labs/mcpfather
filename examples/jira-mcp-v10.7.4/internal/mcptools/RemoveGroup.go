package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the RemoveGroup tool
const RemoveGroupInputSchema = "{\n  \"properties\": {\n    \"groupname\": {\n      \"description\": \"The name of the group to delete.\",\n      \"type\": \"string\"\n    },\n    \"swapGroup\": {\n      \"description\": \"A different group to transfer the restrictions to.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupname\"\n  ],\n  \"type\": \"object\"\n}"

// NewRemoveGroupMCPTool creates the MCP Tool instance for RemoveGroup
func NewRemoveGroupMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RemoveGroup",
		"Delete a specified group - Deletes a group by given group parameter",
		[]byte(RemoveGroupInputSchema),
	)
}

// RemoveGroupHandler is the handler function for the RemoveGroup tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RemoveGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/group", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RemoveGroup"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
