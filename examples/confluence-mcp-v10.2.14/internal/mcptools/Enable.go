package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Enable tool
const EnableInputSchema = "{\n  \"properties\": {\n    \"username\": {\n      \"description\": \"the username identifying the given user\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Enable tool (Status: 401, Content-Type: application/json)
const EnableResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> The calling user is not authenticated or does not have the <b>LICENSED</b> permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Enable tool (Status: 403, Content-Type: application/json)
const EnableResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> The calling user does not have permission to perform this action.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Enable tool (Status: 404, Content-Type: application/json)
const EnableResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> No User exists for the provided username.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewEnableMCPTool creates the MCP Tool instance for Enable
func NewEnableMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Enable",
		"Enable user - Enables the given User identified by username. This method is idempotent i.e. if the user is already enabled then no action will be taken.",
		[]byte(EnableInputSchema),
	)
}

// EnableHandler is the handler function for the Enable tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func EnableHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/admin/user/{username}/enable", args, []string{"username"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Enable"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
