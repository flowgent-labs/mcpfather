package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the Merge tool
const MergeInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The version that will be merged to version moveIssuesTo and removed\",\n      \"type\": \"string\"\n    },\n    \"moveIssuesTo\": {\n      \"description\": \"The version to set fixVersion to on issues where the deleted version is the fix version, If null then the fixVersion is removed.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"moveIssuesTo\"\n  ],\n  \"type\": \"object\"\n}"

// NewMergeMCPTool creates the MCP Tool instance for Merge
func NewMergeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Merge",
		"Merge versions - Merge versions",
		[]byte(MergeInputSchema),
	)
}

// MergeHandler is the handler function for the Merge tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MergeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/version/{id}/mergeto/{moveIssuesTo}", args, []string{"id", "moveIssuesTo"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Merge"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
