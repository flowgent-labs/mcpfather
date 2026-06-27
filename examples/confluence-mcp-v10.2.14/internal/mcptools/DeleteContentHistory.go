package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the DeleteContentHistory tool
const DeleteContentHistoryInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"versionNumber\": {\n      \"description\": \"version number starts from 1 up to current version.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"versionNumber\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the DeleteContentHistory tool (Status: 400, Content-Type: application/json)
const DeleteContentHistoryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if versionNumber is less than 1, does not exist or has already been deleted.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the DeleteContentHistory tool (Status: 403, Content-Type: application/json)
const DeleteContentHistoryResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling user doesn't have permission to edit the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the DeleteContentHistory tool (Status: 404, Content-Type: application/json)
const DeleteContentHistoryResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if the contentId cannot be found.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDeleteContentHistoryMCPTool creates the MCP Tool instance for DeleteContentHistory
func NewDeleteContentHistoryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteContentHistory",
		"Delete content history - Delete a historical version of a page or a blogpost. Current user must have edit permission on content, or it will throw a permission exception.",
		[]byte(DeleteContentHistoryInputSchema),
	)
}

// DeleteContentHistoryHandler is the handler function for the DeleteContentHistory tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteContentHistoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/content/{id}/version/{versionNumber}", args, []string{"id", "versionNumber"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteContentHistory"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
