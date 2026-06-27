package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DefaultColumns tool
const DefaultColumnsInputSchema = "{\n  \"properties\": {\n    \"username\": {\n      \"description\": \"username\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the DefaultColumns tool (Status: 200, Content-Type: application/json)
const DefaultColumnsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of columns for configured for the given user\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewDefaultColumnsMCPTool creates the MCP Tool instance for DefaultColumns
func NewDefaultColumnsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DefaultColumns",
		"Get default columns for user - Returns the default columns for the given user. Admin permission will be required to get columns for a user other than the currently logged in user.",
		[]byte(DefaultColumnsInputSchema),
	)
}

// DefaultColumnsHandler is the handler function for the DefaultColumns tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DefaultColumnsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/columns", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DefaultColumns"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
