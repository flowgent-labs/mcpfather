package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the ResetOrder tool
const ResetOrderInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The sort direction for ordering the list alphabetically. Acceptable values are 'asc' or 'desc' (case-insensitive).\",\n      \"properties\": {\n        \"direction\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ResetOrder tool (Status: 200, Content-Type: application/json)
const ResetOrderResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns all issue link types in reset order.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewResetOrderMCPTool creates the MCP Tool instance for ResetOrder
func NewResetOrderMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ResetOrder",
		"Reset the order of issue link types alphabetically. - Resets the order of issue link types alphabetically.",
		[]byte(ResetOrderInputSchema),
	)
}

// ResetOrderHandler is the handler function for the ResetOrder tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ResetOrderHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issueLinkType/order", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ResetOrder"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
