package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteIssueLinkType tool
const DeleteIssueLinkTypeInputSchema = "{\n  \"properties\": {\n    \"issueLinkTypeId\": {\n      \"description\": \"The issue link type id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueLinkTypeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteIssueLinkTypeMCPTool creates the MCP Tool instance for DeleteIssueLinkType
func NewDeleteIssueLinkTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteIssueLinkType",
		"Delete the specified issue link type - Delete the specified issue link type.",
		[]byte(DeleteIssueLinkTypeInputSchema),
	)
}

// DeleteIssueLinkTypeHandler is the handler function for the DeleteIssueLinkType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteIssueLinkTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issueLinkType/{issueLinkTypeId}", args, []string{"issueLinkTypeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteIssueLinkType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
