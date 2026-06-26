package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetLatestInvocation tool
const GetLatestInvocationInputSchema = "{\n  \"properties\": {\n    \"event\": {\n      \"description\": \"the string id of a specific event to retrieve the last invocation for.\",\n      \"type\": \"string\"\n    },\n    \"outcome\": {\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"outcomes\": {\n      \"description\": \"the outcome to filter for. Can be SUCCESS, FAILURE, ERROR. None specified means that the all will be considered.\",\n      \"type\": \"string\"\n    },\n    \"webhookId\": {\n      \"description\": \"id of the webhook\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"webhookId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetLatestInvocation tool (Status: 200, Content-Type: application/json)
const GetLatestInvocationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a webhook invocation dataset.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetLatestInvocation tool (Status: 401, Content-Type: application/json)
const GetLatestInvocationResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> returned if the currently authenticated user has insufficient permissions to get webhook.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetLatestInvocation tool (Status: 404, Content-Type: application/json)
const GetLatestInvocationResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> returned if the webhook does not exist.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetLatestInvocationMCPTool creates the MCP Tool instance for GetLatestInvocation
func NewGetLatestInvocationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLatestInvocation",
		"Get latest invocations - Get the latest invocations for a specific webhook. The authenticated user must be an administrator to call this resource.",
		[]byte(GetLatestInvocationInputSchema),
	)
}

// GetLatestInvocationHandler is the handler function for the GetLatestInvocation tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLatestInvocationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/webhooks/{webhookId}/latest", args, []string{"webhookId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetLatestInvocation"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
