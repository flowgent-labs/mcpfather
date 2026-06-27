package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UpdateWebhook tool
const UpdateWebhookInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"the representation of the updated values for the webhook\"\n    },\n    \"webhookId\": {\n      \"description\": \"the existing webhook id\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"webhookId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateWebhook tool (Status: 200, Content-Type: application/json)
const UpdateWebhookResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a webhook.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateWebhook tool (Status: 401, Content-Type: application/json)
const UpdateWebhookResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> returned if the currently authenticated user has insufficient permissions to update a webhook.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateWebhook tool (Status: 404, Content-Type: application/json)
const UpdateWebhookResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> returned if the webhook does not exist.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateWebhookMCPTool creates the MCP Tool instance for UpdateWebhook
func NewUpdateWebhookMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateWebhook",
		" Update webhook -  Update an existing webhook. The authenticated user must be an administrator to call this resource.",
		[]byte(UpdateWebhookInputSchema),
	)
}

// UpdateWebhookHandler is the handler function for the UpdateWebhook tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateWebhookHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/webhooks/{webhookId}", args, []string{"webhookId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateWebhook"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
