package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Delete5 tool
const Delete5InputSchema = "{\n  \"properties\": {\n    \"spaceKey\": {\n      \"description\": \"the key of the space to update.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Delete5 tool (Status: 202, Content-Type: application/json)
const Delete5ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 202\n\n**Content-Type:** application/json\n\n> Returns a pointer to the status of the space-deletion task.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Delete5 tool (Status: 404, Content-Type: application/json)
const Delete5ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given key, or if the calling user does not have permission to delete it.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDelete5MCPTool creates the MCP Tool instance for Delete5
func NewDelete5MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Delete5",
		"Delete Space - Deletes a Space. The space is deleted in a long running task, so the space cannot be considered deleted when this resource returns. Clients can follow the status link in the response and poll it until the task completes.",
		[]byte(Delete5InputSchema),
	)
}

// Delete5Handler is the handler function for the Delete5 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Delete5Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/space/{spaceKey}", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Delete5"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
