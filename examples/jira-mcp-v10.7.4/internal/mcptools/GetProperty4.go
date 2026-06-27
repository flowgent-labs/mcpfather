package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProperty4 tool
const GetProperty4InputSchema = "{\n  \"properties\": {\n    \"issueTypeId\": {\n      \"description\": \"The issue type from which the property will be returned.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the property to return.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueTypeId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProperty4 tool (Status: 200, Content-Type: application/json)
const GetProperty4ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the value of the property with a given key from the issue type.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **value** (Type: string):\n      - Example: '{\"hipchat.room.id\":\"support-123\",\"support.time\":\"1m\"}'\n  - **key** (Type: string):\n      - Example: 'issue.support'\n"

// NewGetProperty4MCPTool creates the MCP Tool instance for GetProperty4
func NewGetProperty4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProperty4",
		"Get value of specified issue type's property - Returns the value of the property with a given key from the issue type identified by the id",
		[]byte(GetProperty4InputSchema),
	)
}

// GetProperty4Handler is the handler function for the GetProperty4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProperty4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuetype/{issueTypeId}/properties/{propertyKey}", args, []string{"issueTypeId", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProperty4"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
