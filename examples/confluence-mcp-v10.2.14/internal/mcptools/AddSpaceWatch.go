package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the AddSpaceWatch tool
const AddSpaceWatchInputSchema = "{\n  \"properties\": {\n    \"contentType\": {\n      \"description\": \"the optional content type to delete the watcher for.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"key\": {\n      \"description\": \"userKey of the user to create the new watcher for.\"\n    },\n    \"spaceKey\": {\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"userName of the user to create the new watcher for.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddSpaceWatch tool (Status: 200, Content-Type: application/json)
const AddSpaceWatchResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the watcher was successfully created\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the AddSpaceWatch tool (Status: 404, Content-Type: application/json)
const AddSpaceWatchResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if no content exists for the specified space key or the calling user does not have permission to perform the operation\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewAddSpaceWatchMCPTool creates the MCP Tool instance for AddSpaceWatch
func NewAddSpaceWatchMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddSpaceWatch",
		"Add space watcher - Create a new watcher for the given user and space key. User is optional. If not specified, currently logged-in user will be used. Otherwise, it can be specified by either user key or username. When a user is specified and is different from the logged-in user, the logged-in user needs to be a Confluence administrator. \n\n Example request URI(s):\n\n"+"\x60"+"http://example.com/confluence/rest/api/user/watch/space/SPACEKEY"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/watch/space/SPACEKEY?username=jblogs"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/watch/space/SPACEKEY?key=ff8080815a58e24c015a58e263710000"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/watch/space/SPACEKEY?contentType=blogpost"+"\x60"+"",
		[]byte(AddSpaceWatchInputSchema),
	)
}

// AddSpaceWatchHandler is the handler function for the AddSpaceWatch tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddSpaceWatchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/user/watch/space/{spaceKey}", args, []string{"spaceKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "AddSpaceWatch"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
