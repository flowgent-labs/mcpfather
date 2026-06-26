package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UpdateColorScheme tool
const UpdateColorSchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"New color scheme\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the UpdateColorScheme tool (Status: 200, Content-Type: application/json)
const UpdateColorSchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of the updated color scheme\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateColorScheme tool (Status: 400, Content-Type: application/json)
const UpdateColorSchemeResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if there are invalid colors in the request\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateColorScheme tool (Status: 403, Content-Type: application/json)
const UpdateColorSchemeResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the user is not a site admin\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateColorSchemeMCPTool creates the MCP Tool instance for UpdateColorScheme
func NewUpdateColorSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateColorScheme",
		"Set global color scheme - Update the current color scheme of the instance\n\n",
		[]byte(UpdateColorSchemeInputSchema),
	)
}

// UpdateColorSchemeHandler is the handler function for the UpdateColorScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateColorSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/color-scheme", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateColorScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
