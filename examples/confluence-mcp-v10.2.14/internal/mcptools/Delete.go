package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Delete tool
const DeleteInputSchema = "{\n  \"properties\": {\n    \"groupName\": {\n      \"description\": \"the group name to be deleted\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupName\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Delete tool (Status: 400, Content-Type: application/json)
const DeleteResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if user is attempting to delete the last admin group\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Delete tool (Status: 403, Content-Type: application/json)
const DeleteResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> returned if user does not have correct permission\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Delete tool (Status: 404, Content-Type: application/json)
const DeleteResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> returned if group cannot be found\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDeleteMCPTool creates the MCP Tool instance for Delete
func NewDeleteMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Delete",
		"Delete group - Deletes the given group identified by name.",
		[]byte(DeleteInputSchema),
	)
}

// DeleteHandler is the handler function for the Delete tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/admin/group/{groupName}", args, []string{"groupName"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Delete"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
