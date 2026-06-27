package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPropertiesKeys_8a84b00e tool
const GetPropertiesKeys_8a84b00eInputSchema = "{\n  \"properties\": {\n    \"dashboardId\": {\n      \"description\": \"The dashboard id.\",\n      \"type\": \"string\"\n    },\n    \"itemId\": {\n      \"description\": \"The dashboard item from which keys will be returned.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"dashboardId\",\n    \"itemId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPropertiesKeys_8a84b00e tool (Status: 200, Content-Type: application/json)
const GetPropertiesKeys_8a84b00eResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the dashboard item was found.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **keys** (Type: array):\n    - **Items** (Type: object):\n      - **key** (Type: string):\n          - Example: 'issue.support'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/EX-2/properties/issue.support'\n"

// NewGetPropertiesKeys_8a84b00eMCPTool creates the MCP Tool instance for GetPropertiesKeys_8a84b00e
func NewGetPropertiesKeys_8a84b00eMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPropertiesKeys_8a84b00e",
		"Get all properties keys for a dashboard item - Returns the keys of all properties for the dashboard item identified by the id.",
		[]byte(GetPropertiesKeys_8a84b00eInputSchema),
	)
}

// GetPropertiesKeys_8a84b00eHandler is the handler function for the GetPropertiesKeys_8a84b00e tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPropertiesKeys_8a84b00eHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/dashboard/{dashboardId}/items/{itemId}/properties", args, []string{"dashboardId", "itemId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPropertiesKeys_8a84b00e"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
