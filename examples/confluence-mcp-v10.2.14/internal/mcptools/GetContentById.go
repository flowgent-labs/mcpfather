package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetContentById tool
const GetContentByIdInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"A comma separated list of properties to expand on the content. Default value: \\u003ccode\\u003ehistory,space,version\\u003c/code\\u003e. \\n\\n We can also specify some extensions such as \\u003ccode\\u003eextensions.inlineProperties\\u003c/code\\u003e (for getting inline comment-specific properties) or \\u003ccode\\u003eextensions.resolution\\u003c/code\\u003e for the resolution status of each comment in the results\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"the id of the content.\",\n      \"type\": \"string\"\n    },\n    \"status\": {\n      \"description\": \"list of Content statuses to filter results on. \\n\\n Default value: \\u003ccode\\u003e[current]\\u003c/code\\u003e.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"version\": {\n      \"description\": \"version of the content.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"xoauth_requestor_id\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetContentById tool (Status: 200, Content-Type: application/json)
const GetContentByIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns  a JSON representation of the content, or a 404 NOT FOUND if there is no content with the given id or if the user is not permitted.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetContentById tool (Status: 404, Content-Type: application/json)
const GetContentByIdResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetContentByIdMCPTool creates the MCP Tool instance for GetContentById
func NewGetContentByIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetContentById",
		"Get content by ID - Returns a piece of Content. Example request URI(s): \n\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234?expand=space,body.view,version,container"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234?status=any"+"\x60"+"",
		[]byte(GetContentByIdInputSchema),
	)
}

// GetContentByIdHandler is the handler function for the GetContentById tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetContentByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetContentById"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
