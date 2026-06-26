package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateIssueLinkType tool
const UpdateIssueLinkTypeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"All information about the link relationship.\",\n      \"properties\": {\n        \"id\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"inward\": {\n          \"example\": \"is duplicated by\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"Duplicate\",\n          \"type\": \"string\"\n        },\n        \"outward\": {\n          \"example\": \"duplicates\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/issueLinkType/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueLinkTypeId\": {\n      \"description\": \"The issue link type id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"issueLinkTypeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateIssueLinkTypeMCPTool creates the MCP Tool instance for UpdateIssueLinkType
func NewUpdateIssueLinkTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateIssueLinkType",
		"Update the specified issue link type - Update the specified issue link type.",
		[]byte(UpdateIssueLinkTypeInputSchema),
	)
}

// UpdateIssueLinkTypeHandler is the handler function for the UpdateIssueLinkType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateIssueLinkTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issueLinkType/{issueLinkTypeId}", args, []string{"issueLinkTypeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateIssueLinkType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
