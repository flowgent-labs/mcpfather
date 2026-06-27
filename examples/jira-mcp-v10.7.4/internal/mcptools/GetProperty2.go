package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProperty2 tool
const GetProperty2InputSchema = "{\n  \"properties\": {\n    \"commentId\": {\n      \"description\": \"the comment from which the property will be returned.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"the key of the property to return.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"commentId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProperty2 tool (Status: 200, Content-Type: application/json)
const GetProperty2ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the value of the property with a given key from the comment.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **value** (Type: string):\n      - Example: '{\"hipchat.room.id\":\"support-123\",\"support.time\":\"1m\"}'\n  - **key** (Type: string):\n      - Example: 'issue.support'\n"

// NewGetProperty2MCPTool creates the MCP Tool instance for GetProperty2
func NewGetProperty2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProperty2",
		"Get a property from a comment - Returns the value of the property with a given key from the comment identified by the key or by the id. The user who retrieves the property is required to have permissions to read the comment.",
		[]byte(GetProperty2InputSchema),
	)
}

// GetProperty2Handler is the handler function for the GetProperty2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProperty2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/comment/{commentId}/properties/{propertyKey}", args, []string{"commentId", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProperty2"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
