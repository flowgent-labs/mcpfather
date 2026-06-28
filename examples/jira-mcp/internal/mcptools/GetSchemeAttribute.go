package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetSchemeAttribute tool
const GetSchemeAttributeInputSchema = "{\n  \"properties\": {\n    \"attributeKey\": {\n      \"description\": \"The key of the permission scheme attribute.\",\n      \"type\": \"string\"\n    },\n    \"permissionSchemeId\": {\n      \"description\": \"The id of the permission scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"attributeKey\",\n    \"permissionSchemeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetSchemeAttribute tool (Status: 200, Content-Type: application/json)
const GetSchemeAttributeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Permission scheme attribute\n\n## Response Structure\n\n- Structure (Type: object):\n  - **value** (Type: string):\n  - **key** (Type: string):\n"

// NewGetSchemeAttributeMCPTool creates the MCP Tool instance for GetSchemeAttribute
func NewGetSchemeAttributeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSchemeAttribute",
		"Get scheme attribute by key - Returns the attribute for a permission scheme specified by permission scheme id and attribute key.",
		[]byte(GetSchemeAttributeInputSchema),
	)
}

// GetSchemeAttributeHandler is the handler function for the GetSchemeAttribute tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSchemeAttributeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/permissionscheme/{permissionSchemeId}/attribute/{attributeKey}", args, []string{"attributeKey", "permissionSchemeId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSchemeAttribute")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
