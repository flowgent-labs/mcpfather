package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Delete4 tool
const Delete4InputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"the key of the property.\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the key of the space.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Delete4 tool (Status: 404, Content-Type: application/json)
const Delete4ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the give key, property with the given property key, or if the calling user does not have permission to view the space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDelete4MCPTool creates the MCP Tool instance for Delete4
func NewDelete4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Delete4",
		"Delete space property - Deletes a space property. \n\nExample request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/api/space/TST/property/example-property-key?expand=space,version"+"\x60"+"",
		[]byte(Delete4InputSchema),
	)
}

// Delete4Handler is the handler function for the Delete4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Delete4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/space/{spaceKey}/property/{key}", args, []string{"key", "spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Delete4"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
