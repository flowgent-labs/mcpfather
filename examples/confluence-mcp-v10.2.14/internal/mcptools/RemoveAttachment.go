package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the RemoveAttachment tool
const RemoveAttachmentInputSchema = "{\n  \"properties\": {\n    \"attachmentId\": {\n      \"description\": \"the id of the attachment to be removed.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"The id of the content the attachment is on.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"attachmentId\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the RemoveAttachment tool (Status: 400, Content-Type: application/json)
const RemoveAttachmentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if the user does not have permission to remove the attachment.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the RemoveAttachment tool (Status: 404, Content-Type: application/json)
const RemoveAttachmentResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n>  Returned if the specified attachment does not exist.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewRemoveAttachmentMCPTool creates the MCP Tool instance for RemoveAttachment
func NewRemoveAttachmentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RemoveAttachment",
		"Remove attachment - This method will delete the attachment identified by attachmentId.\n\nIt returns a boolean response indicating whether the operation was successful or not. If the specified attachment or version does not exist, or if the user does not have permission to remove the attachment, appropriate exceptions are thrown and mapped to their corresponding HTTP responses.",
		[]byte(RemoveAttachmentInputSchema),
	)
}

// RemoveAttachmentHandler is the handler function for the RemoveAttachment tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RemoveAttachmentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/content/{id}/child/attachment/{attachmentId}", args, []string{"attachmentId", "id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RemoveAttachment"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
