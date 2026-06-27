package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the TestWebhook tool
const TestWebhookInputSchema = "{\n  \"properties\": {\n    \"url\": {\n      \"description\": \"the url in which to connect to\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"url\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the TestWebhook tool (Status: 401, Content-Type: application/json)
const TestWebhookResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> returned if the currently authenticated user has insufficient permissions to test a connection.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the TestWebhook tool (Status: 404, Content-Type: application/json)
const TestWebhookResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> returned if repository does not exist.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewTestWebhookMCPTool creates the MCP Tool instance for TestWebhook
func NewTestWebhookMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"TestWebhook",
		"Test webhook - Test connectivity to a specific endpoint. The authenticated user must be an administrator to call this resource.",
		[]byte(TestWebhookInputSchema),
	)
}

// TestWebhookHandler is the handler function for the TestWebhook tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func TestWebhookHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/webhooks/test", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "TestWebhook"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
