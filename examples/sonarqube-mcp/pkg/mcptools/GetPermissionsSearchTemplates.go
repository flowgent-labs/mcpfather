package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetPermissionsSearchTemplates tool
const GetPermissionsSearchTemplatesInputSchema = "{\n  \"properties\": {\n    \"p\": {\n      \"description\": \"1-based page number\",\n      \"type\": \"string\"\n    },\n    \"ps\": {\n      \"description\": \"Page size. Must be greater than or equal to 0 and less or equal than 500. If this and p param are not provided, pagination is disabled and all results are returned. If pageSize=0, no results are returned but the response will contain the total count of matching templates.\",\n      \"type\": \"string\"\n    },\n    \"q\": {\n      \"description\": \"Limit search to permission template names that contain the supplied string.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPermissionsSearchTemplates tool (Status: 200, Content-Type: application/json)
const GetPermissionsSearchTemplatesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetPermissionsSearchTemplatesMCPTool creates the MCP Tool instance for GetPermissionsSearchTemplates
func NewGetPermissionsSearchTemplatesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPermissionsSearchTemplates",
		"List permission templates. - List permission templates. Requires the following permission: 'Administer System'.",
		[]byte(GetPermissionsSearchTemplatesInputSchema),
	)
}

// GetPermissionsSearchTemplatesHandler is the handler function for the GetPermissionsSearchTemplates tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPermissionsSearchTemplatesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/permissions/search_templates", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPermissionsSearchTemplates")
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
