package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetRefinedVelocity tool
const GetRefinedVelocityInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"boardId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRefinedVelocity tool (Status: 200, Content-Type: application/json)
const GetRefinedVelocityResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the board exists and the property was found.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **value** (Type: boolean):\n      - Example: 'true'\n"

// NewGetRefinedVelocityMCPTool creates the MCP Tool instance for GetRefinedVelocity
func NewGetRefinedVelocityMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRefinedVelocity",
		"Get the value of the refined velocity setting - Returns the value of the setting for refined velocity chart",
		[]byte(GetRefinedVelocityInputSchema),
	)
}

// GetRefinedVelocityHandler is the handler function for the GetRefinedVelocity tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRefinedVelocityHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/settings/refined-velocity", args, []string{"boardId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetRefinedVelocity"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
