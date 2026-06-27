package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateShowWhenEmptyIndicator tool
const UpdateShowWhenEmptyIndicatorInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"id of field\",\n      \"type\": \"string\"\n    },\n    \"newValue\": {\n      \"description\": \"new value of 'showWhenEmptyIndicator'\",\n      \"type\": \"boolean\"\n    },\n    \"screenId\": {\n      \"description\": \"id of screen\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"tabId\": {\n      \"description\": \"id of tab\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"newValue\",\n    \"screenId\",\n    \"tabId\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateShowWhenEmptyIndicatorMCPTool creates the MCP Tool instance for UpdateShowWhenEmptyIndicator
func NewUpdateShowWhenEmptyIndicatorMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateShowWhenEmptyIndicator",
		"Update 'showWhenEmptyIndicator' for a field - Update 'showWhenEmptyIndicator' for given field on screen.",
		[]byte(UpdateShowWhenEmptyIndicatorInputSchema),
	)
}

// UpdateShowWhenEmptyIndicatorHandler is the handler function for the UpdateShowWhenEmptyIndicator tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateShowWhenEmptyIndicatorHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/screens/{screenId}/tabs/{tabId}/fields/{id}/updateShowWhenEmptyIndicator/{newValue}", args, []string{"id", "newValue", "screenId", "tabId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateShowWhenEmptyIndicator"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
