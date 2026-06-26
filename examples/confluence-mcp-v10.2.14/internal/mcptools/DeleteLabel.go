package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the DeleteLabel tool
const DeleteLabelInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"label\": {\n      \"description\": \"the name of the label to be removed from the content\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"label\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the DeleteLabel tool (Status: 400, Content-Type: application/json)
const DeleteLabelResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if trying to delete a label with \"/\" in the name.permission to the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the DeleteLabel tool (Status: 403, Content-Type: application/json)
const DeleteLabelResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if user has view permission, but no edit permission to the content.permission to the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the DeleteLabel tool (Status: 404, Content-Type: application/json)
const DeleteLabelResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if content or label doesn't exist, or calling user doesn't have view permission to the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDeleteLabelMCPTool creates the MCP Tool instance for DeleteLabel
func NewDeleteLabelMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteLabel",
		"Delete label - Deletes a labels to the specified content. The body is the json representation of the list. When calling this method through REST the label parameter doesn't accept "+"\x60"+"/"+"\x60"+" characters in label names, because of security constraints. For this case please use the query parameter version of this method ("+"\x60"+"/content/{id}/label?name={label}"+"\x60"+")",
		[]byte(DeleteLabelInputSchema),
	)
}

// DeleteLabelHandler is the handler function for the DeleteLabel tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteLabelHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/content/{id}/label/{label}", args, []string{"id", "label"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteLabel"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
