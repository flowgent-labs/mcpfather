package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Update tool
const UpdateInputSchema = "{\n  \"properties\": {\n    \"attachmentId\": {\n      \"description\": \"the id of the attachment to update.\",\n      \"type\": \"string\"\n    },\n    \"body\": {},\n    \"id\": {\n      \"description\": \"The id of the content the attachment is on.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"attachmentId\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Update tool (Status: 200, Content-Type: application/json)
const UpdateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a JSON representation of the attachment after being updated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update tool (Status: 400, Content-Type: application/json)
const UpdateResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  Returned if the attachment id or the attachment version number are invalid.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update tool (Status: 403, Content-Type: application/json)
const UpdateResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if you are not permitted to update or move the attachment to a different container.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update tool (Status: 404, Content-Type: application/json)
const UpdateResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n>  Returned if no attachment is found for the attachmentId.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update tool (Status: 409, Content-Type: application/json)
const UpdateResponseTemplate_E = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Returned if the version of the supplied Attachment does not match the exact version of the Attachment stored in the database.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateMCPTool creates the MCP Tool instance for Update
func NewUpdateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Update",
		"Update non-binary data of an Attachment - Update the non-binary data of an attachment.This resource can be used to update an attachment's filename, media-type, comment, and parent container.",
		[]byte(UpdateInputSchema),
	)
}

// UpdateHandler is the handler function for the Update tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/content/{id}/child/attachment/{attachmentId}", args, []string{"attachmentId", "id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Update"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
