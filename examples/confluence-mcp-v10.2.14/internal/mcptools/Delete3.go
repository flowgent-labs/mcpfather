package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Delete3 tool
const Delete3InputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"  the id of the content.\",\n      \"type\": \"string\"\n    },\n    \"status\": {\n      \"description\": \"the status of the content to be deleted.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Delete3 tool (Status: 404, Content-Type: application/json)
const Delete3ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to trash or purge the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Delete3 tool (Status: 409, Content-Type: application/json)
const Delete3ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Returned if there is a stale data object conflict when trying to delete a draft.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDelete3MCPTool creates the MCP Tool instance for Delete3
func NewDelete3MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Delete3",
		"Delete content - Trashes or purges a piece of Content, based on its ContentType and ContentStatus. \n\nThere are three cases:\n\n- If the content is trashable and its status is current, it will be trashed.\n\n- If the content is trashable, its status is trashed and the status query parameter in the request is trashed, the content will be purged from the trash and deleted permanently.\n\n- If the content is not trashable it will be deleted permanently without being trashed.",
		[]byte(Delete3InputSchema),
	)
}

// Delete3Handler is the handler function for the Delete3 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Delete3Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/content/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Delete3"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
