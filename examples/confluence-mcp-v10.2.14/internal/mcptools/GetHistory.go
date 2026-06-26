package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetHistory tool
const GetHistoryInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the content. Default value: \\u003ccode\\u003epreviousVersion,nextVersion,lastUpdated\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"  the id of the content.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetHistory tool (Status: 200, Content-Type: application/json)
const GetHistoryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of the content's history\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetHistory tool (Status: 404, Content-Type: application/json)
const GetHistoryResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetHistoryMCPTool creates the MCP Tool instance for GetHistory
func NewGetHistoryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetHistory",
		"Get history of content - Returns the history of a particular piece of content. Example request URI(s): \n\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/history"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/history?expand=previousVersion,nextVersion,lastUpdated"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/history?cql=creator=currentUser()&cqlcontext={\"spaceKey\":\"TST\", \"contentId\":\"55\"}&expand=previousVersion,nextVersion,lastUpdated"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/history?cql=creator=currentUser()&cqlcontext={\"spaceKey\":\"TST\", \"contentId\":\"55\"}&expand=previousVersion,nextVersion,lastUpdated&start=0&limit=10"+"\x60"+"",
		[]byte(GetHistoryInputSchema),
	)
}

// GetHistoryHandler is the handler function for the GetHistory tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetHistoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/history", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetHistory"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
