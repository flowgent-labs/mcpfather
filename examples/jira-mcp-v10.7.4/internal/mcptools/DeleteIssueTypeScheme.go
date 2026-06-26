package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteIssueTypeScheme tool
const DeleteIssueTypeSchemeInputSchema = "{\n  \"properties\": {\n    \"schemeId\": {\n      \"description\": \"The id of the issue type scheme to remove.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteIssueTypeSchemeMCPTool creates the MCP Tool instance for DeleteIssueTypeScheme
func NewDeleteIssueTypeSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteIssueTypeScheme",
		"Delete specified issue type scheme - Deletes the specified issue type scheme. Any projects associated with this IssueTypeScheme will be automatically associated with the global default IssueTypeScheme.",
		[]byte(DeleteIssueTypeSchemeInputSchema),
	)
}

// DeleteIssueTypeSchemeHandler is the handler function for the DeleteIssueTypeScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteIssueTypeSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issuetypescheme/{schemeId}", args, []string{"schemeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteIssueTypeScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
