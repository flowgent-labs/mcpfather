package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the AddTab tool
const AddTabInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"id\": {\n          \"example\": 10000,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"name\": {\n          \"example\": \"Fields Tab\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"screenId\": {\n      \"description\": \"id of screen\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"screenId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddTab tool (Status: 200, Content-Type: application/json)
const AddTabResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns newly created tab.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: integer, int64):\n      - Example: '10000'\n  - **name** (Type: string):\n      - Example: 'Fields Tab'\n"

// NewAddTabMCPTool creates the MCP Tool instance for AddTab
func NewAddTabMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddTab",
		"Create tab for a screen - Creates tab for given screen.",
		[]byte(AddTabInputSchema),
	)
}

// AddTabHandler is the handler function for the AddTab tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddTabHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/screens/{screenId}/tabs", args, []string{"screenId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "AddTab"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
