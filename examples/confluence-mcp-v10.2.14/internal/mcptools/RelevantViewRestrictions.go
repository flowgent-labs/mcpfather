package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the RelevantViewRestrictions tool
const RelevantViewRestrictionsInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"A comma separated list of properties to expand on the content properties. Default value: relevantViewRestrictions\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 50,\n      \"description\": \"pagination limit, Max 50 per page\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"pagination start.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the RelevantViewRestrictions tool (Status: 200, Content-Type: application/json)
const RelevantViewRestrictionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the restrictions group by operations.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewRelevantViewRestrictionsMCPTool creates the MCP Tool instance for RelevantViewRestrictions
func NewRelevantViewRestrictionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RelevantViewRestrictions",
		"Get all view restriction both direct and inherited. - Returns relevant view restriction both direct and inherited for a single content.",
		[]byte(RelevantViewRestrictionsInputSchema),
	)
}

// RelevantViewRestrictionsHandler is the handler function for the RelevantViewRestrictions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RelevantViewRestrictionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/restriction/relevantViewRestrictions", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RelevantViewRestrictions"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
