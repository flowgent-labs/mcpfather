package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the IsWatchingContent tool
const IsWatchingContentInputSchema = "{\n  \"properties\": {\n    \"contentId\": {\n      \"description\": \"id of the content.\",\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"userkey of the user to check for watching state.\"\n    },\n    \"username\": {\n      \"description\": \"username of the user to check for watching state.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"contentId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the IsWatchingContent tool (Status: 404, Content-Type: application/json)
const IsWatchingContentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if no content exists for the specified content id or calling user does not have permission to perform the operation\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewIsWatchingContentMCPTool creates the MCP Tool instance for IsWatchingContent
func NewIsWatchingContentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"IsWatchingContent",
		"Get information about content watcher - Get information about whether a user is watching a specified content. User is optional. If not specified, currently logged-in user will be used. Otherwise, it can be specified by either user key or username. When a user is specified and is different from the logged-in user, the logged-in user needs to be a Confluence administrator. \n\n Example request URI(s):\n\n"+"\x60"+"http://example.com/confluence/rest/api/user/watch/content/131213"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/watch/content/131213?username=jblogs"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/watch/content/131213?key=ff8080815a58e24c015a58e263710000"+"\x60"+"",
		[]byte(IsWatchingContentInputSchema),
	)
}

// IsWatchingContentHandler is the handler function for the IsWatchingContent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func IsWatchingContentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/user/watch/content/{contentId}", args, []string{"contentId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "IsWatchingContent"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
