package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostApplicationsCreate tool
const PostApplicationsCreateInputSchema = "{\n  \"properties\": {\n    \"description\": {\n      \"description\": \"Application description\",\n      \"maxLength\": 256,\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"Application key. A suitable key will be generated if not provided\",\n      \"maxLength\": 400,\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"Application name\",\n      \"maxLength\": 256,\n      \"type\": \"string\"\n    },\n    \"visibility\": {\n      \"description\": \"Whether the created application should be visible to everyone, or only specific user/groups. If no visibility is specified, the default visibility will be used.\",\n      \"enum\": [\n        \"private\",\n        \"public\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"name\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PostApplicationsCreate tool (Status: 200, Content-Type: application/json)
const PostApplicationsCreateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewPostApplicationsCreateMCPTool creates the MCP Tool instance for PostApplicationsCreate
func NewPostApplicationsCreateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostApplicationsCreate",
		"Create a new application. - Create a new application. Requires 'Administer System' permission or 'Create Applications' permission",
		[]byte(PostApplicationsCreateInputSchema),
	)
}

// PostApplicationsCreateHandler is the handler function for the PostApplicationsCreate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostApplicationsCreateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/applications/create", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostApplicationsCreate")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
