package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProperty1_57562797 tool
const GetProperty1_57562797InputSchema = "{\n  \"properties\": {\n    \"dashboardId\": {\n      \"description\": \"The dashboard id.\",\n      \"type\": \"string\"\n    },\n    \"itemId\": {\n      \"description\": \"The dashboard item from which the property will be returned.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the property to return.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"dashboardId\",\n    \"itemId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProperty1_57562797 tool (Status: 200, Content-Type: application/json)
const GetProperty1_57562797ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the dashboard item property was found.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **value** (Type: string):\n      - Example: '{\"hipchat.room.id\":\"support-123\",\"support.time\":\"1m\"}'\n  - **key** (Type: string):\n      - Example: 'issue.support'\n"

// NewGetProperty1_57562797MCPTool creates the MCP Tool instance for GetProperty1_57562797
func NewGetProperty1_57562797MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProperty1_57562797",
		"Get a property from a dashboard item - Returns the value of the property with a given key from the dashboard item identified by the id.",
		[]byte(GetProperty1_57562797InputSchema),
	)
}

// GetProperty1_57562797Handler is the handler function for the GetProperty1_57562797 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProperty1_57562797Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/dashboard/{dashboardId}/items/{itemId}/properties/{propertyKey}", args, []string{"dashboardId", "itemId", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProperty1_57562797"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
