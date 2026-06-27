package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the DeleteLabelWithQueryParam tool
const DeleteLabelWithQueryParamInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"the name of the label to be removed from the content\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the DeleteLabelWithQueryParam tool (Status: 403, Content-Type: application/json)
const DeleteLabelWithQueryParamResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n>  Returned if user has view permission, but no edit permission to the content.permission to the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the DeleteLabelWithQueryParam tool (Status: 404, Content-Type: application/json)
const DeleteLabelWithQueryParamResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if content or label doesn't exist, or calling user doesn't have view permission to the content.permission to the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDeleteLabelWithQueryParamMCPTool creates the MCP Tool instance for DeleteLabelWithQueryParam
func NewDeleteLabelWithQueryParamMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteLabelWithQueryParam",
		"Delete label with query param - Deletes a labels to the specified content.",
		[]byte(DeleteLabelWithQueryParamInputSchema),
	)
}

// DeleteLabelWithQueryParamHandler is the handler function for the DeleteLabelWithQueryParam tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteLabelWithQueryParamHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/content/{id}/label", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteLabelWithQueryParam"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
