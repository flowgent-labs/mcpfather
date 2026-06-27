package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UpdateColorSchemeType tool
const UpdateColorSchemeTypeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"New color scheme type\"\n    },\n    \"spaceKey\": {\n      \"description\": \"space key of the space to update color scheme type for.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateColorSchemeType tool (Status: 200, Content-Type: application/json)
const UpdateColorSchemeTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the updated color scheme type\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateColorSchemeType tool (Status: 400, Content-Type: application/json)
const UpdateColorSchemeTypeResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if space type is invalid in the request\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateColorSchemeType tool (Status: 403, Content-Type: application/json)
const UpdateColorSchemeTypeResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if a space with the user is not space admin for the given space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateColorSchemeType tool (Status: 404, Content-Type: application/json)
const UpdateColorSchemeTypeResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if a space with the given space key does not exist.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateColorSchemeTypeMCPTool creates the MCP Tool instance for UpdateColorSchemeType
func NewUpdateColorSchemeTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateColorSchemeType",
		"Update Space color scheme type - Update the color scheme type used for a space, currently it can be global or custom\n\n",
		[]byte(UpdateColorSchemeTypeInputSchema),
	)
}

// UpdateColorSchemeTypeHandler is the handler function for the UpdateColorSchemeType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateColorSchemeTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/space/{spaceKey}/color-scheme/type", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateColorSchemeType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
