package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Update4 tool
const Update4InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"the space being updated\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the key of the space to update.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Update4 tool (Status: 200, Content-Type: application/json)
const Update4ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of a space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update4 tool (Status: 404, Content-Type: application/json)
const Update4ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given key, or if the calling userdoes not have permission to update it.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdate4MCPTool creates the MCP Tool instance for Update4
func NewUpdate4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Update4",
		"Update Space - Updates a Space. The incoming Space must include a Key and Name, and should include a Description",
		[]byte(Update4InputSchema),
	)
}

// Update4Handler is the handler function for the Update4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Update4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/space/{spaceKey}", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Update4"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
