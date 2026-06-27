package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Delete1 tool
const Delete1InputSchema = "{\n  \"properties\": {\n    \"username\": {\n      \"description\": \"the username identifying the given user\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Delete1 tool (Status: 202, Content-Type: application/json)
const Delete1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 202\n\n**Content-Type:** application/json\n\n> Produces a HTTP Accept 202 response from some other resource pointing to this class's LongTaskStatus resource.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Delete1 tool (Status: 401, Content-Type: application/json)
const Delete1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> The calling user is not authenticated or does not have the <b>LICENSED</b> permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Delete1 tool (Status: 403, Content-Type: application/json)
const Delete1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> The calling user does not have permission to perform this action.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Delete1 tool (Status: 404, Content-Type: application/json)
const Delete1ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> No User exists for the provided username.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDelete1MCPTool creates the MCP Tool instance for Delete1
func NewDelete1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Delete1",
		"Delete user - Deletes the given User identified by username. This action is processed asynchronously.",
		[]byte(Delete1InputSchema),
	)
}

// Delete1Handler is the handler function for the Delete1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Delete1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/admin/user/{username}", args, []string{"username"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Delete1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
