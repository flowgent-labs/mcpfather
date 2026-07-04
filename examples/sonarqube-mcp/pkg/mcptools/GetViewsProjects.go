package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetViewsProjects tool
const GetViewsProjectsInputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"Portfolio key\",\n      \"type\": \"string\"\n    },\n    \"p\": {\n      \"default\": \"1\",\n      \"description\": \"Index of the page to display\",\n      \"type\": \"string\"\n    },\n    \"ps\": {\n      \"default\": \"100\",\n      \"description\": \"Size for the paging to apply\",\n      \"type\": \"string\"\n    },\n    \"query\": {\n      \"description\": \"If specified, only projects whose key contain the query will be returned\",\n      \"type\": \"string\"\n    },\n    \"selected\": {\n      \"description\": \"Depending on the value, show only selected items (selected=selected), deselected items (selected=deselected), or all items with their selection status (selected=all).\",\n      \"enum\": [\n        \"all\",\n        \"deselected\",\n        \"selected\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetViewsProjects tool (Status: 200, Content-Type: application/json)
const GetViewsProjectsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetViewsProjectsMCPTool creates the MCP Tool instance for GetViewsProjects
func NewGetViewsProjectsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetViewsProjects",
		"List projects manually selected in a portfolio. - List projects manually selected in a portfolio. Requires 'Administrator' permission on the portfolio.",
		[]byte(GetViewsProjectsInputSchema),
	)
}

// GetViewsProjectsHandler is the handler function for the GetViewsProjects tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetViewsProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/views/projects", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetViewsProjects")
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
