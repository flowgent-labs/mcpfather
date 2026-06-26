package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProperty_69ef7bff tool
const GetProperty_69ef7bffInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"boardId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProperty_69ef7bff tool (Status: 200, Content-Type: application/json)
const GetProperty_69ef7bffResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested property.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **keys** (Type: array):\n    - **Items** (Type: object):\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/EX-2/properties/issue.support'\n      - **key** (Type: string):\n          - Example: 'issue.support'\n"

// NewGetProperty_69ef7bffMCPTool creates the MCP Tool instance for GetProperty_69ef7bff
func NewGetProperty_69ef7bffMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProperty_69ef7bff",
		"Get a property from a board - Returns the value of the property with a given key from the board identified by the provided id. The user who retrieves the property is required to have permissions to view the board.",
		[]byte(GetProperty_69ef7bffInputSchema),
	)
}

// GetProperty_69ef7bffHandler is the handler function for the GetProperty_69ef7bff tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProperty_69ef7bffHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/properties/{propertyKey}", args, []string{"boardId", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProperty_69ef7bff"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
