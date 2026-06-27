package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the CreateWebhook tool
const CreateWebhookInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"the webhook to be created.\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateWebhook tool (Status: 201, Content-Type: application/json)
const CreateWebhookResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> returns a created webhook.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateWebhook tool (Status: 400, Content-Type: application/json)
const CreateWebhookResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if The webhook parameters were invalid or not supplied.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateWebhook tool (Status: 401, Content-Type: application/json)
const CreateWebhookResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> The currently authenticated user has insufficient permissions to create webhooks.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreateWebhookMCPTool creates the MCP Tool instance for CreateWebhook
func NewCreateWebhookMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateWebhook",
		"Create webhook - Create a webhook via the URL. The authenticated user must be an administrator to call this resource.",
		[]byte(CreateWebhookInputSchema),
	)
}

// CreateWebhookHandler is the handler function for the CreateWebhook tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateWebhookHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/webhooks", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateWebhook"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
