package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the FindByKey tool
const FindByKeyInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the content properties. Default value: \\u003ccode\\u003eversion\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"the key of the content property. Required.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"description\": \"the limit of the number of labels to return, this may be restricted by fixed system limits\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"key\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the FindByKey tool (Status: 200, Content-Type: application/json)
const FindByKeyResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a JSON representation of the content property.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the FindByKey tool (Status: 404, Content-Type: application/json)
const FindByKeyResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or no property with the given key, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewFindByKeyMCPTool creates the MCP Tool instance for FindByKey
func NewFindByKeyMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FindByKey",
		"Find content property by key - Returns a content property. Example request URI(s): \n\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/property/example-property-key?expand=content,version"+"\x60"+"",
		[]byte(FindByKeyInputSchema),
	)
}

// FindByKeyHandler is the handler function for the FindByKey tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FindByKeyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/property/{key}", args, []string{"id", "key"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "FindByKey")
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
