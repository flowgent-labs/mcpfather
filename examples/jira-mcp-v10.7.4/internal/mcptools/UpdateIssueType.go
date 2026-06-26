package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateIssueType tool
const UpdateIssueTypeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"All information about the issue type.\",\n      \"properties\": {\n        \"avatarId\": {\n          \"example\": 1,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"description\": {\n          \"example\": \"description\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"name\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"The issue type id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateIssueTypeMCPTool creates the MCP Tool instance for UpdateIssueType
func NewUpdateIssueTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateIssueType",
		"Update specified issue type from JSON representation - Updates the specified issue type from a JSON representation.",
		[]byte(UpdateIssueTypeInputSchema),
	)
}

// UpdateIssueTypeHandler is the handler function for the UpdateIssueType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateIssueTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issuetype/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateIssueType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
