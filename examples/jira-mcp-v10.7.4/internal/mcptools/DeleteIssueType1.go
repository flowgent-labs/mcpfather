package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteIssueType1 tool
const DeleteIssueType1InputSchema = "{\n  \"properties\": {\n    \"alternativeIssueTypeId\": {\n      \"description\": \"The id of an issue type to which issues associated with the removed issue type will be migrated.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"The issue type id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"alternativeIssueTypeId\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteIssueType1MCPTool creates the MCP Tool instance for DeleteIssueType1
func NewDeleteIssueType1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteIssueType1",
		"Delete specified issue type and migrate associated issues - Deletes the specified issue type. If the issue type has any associated issues, these issues will be migrated to the alternative issue type specified in the parameter.",
		[]byte(DeleteIssueType1InputSchema),
	)
}

// DeleteIssueType1Handler is the handler function for the DeleteIssueType1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteIssueType1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issuetype/{id}", args, []string{"alternativeIssueTypeId", "id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteIssueType1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
