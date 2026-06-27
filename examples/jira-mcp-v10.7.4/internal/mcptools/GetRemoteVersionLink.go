package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetRemoteVersionLink tool
const GetRemoteVersionLinkInputSchema = "{\n  \"properties\": {\n    \"globalId\": {\n      \"description\": \"The id of the remote issue link to be returned. If (not provided) all remote links for the issue are returned.\\nRemote version links follow the same general rules that Issue Links do, except that they are permitted to\\nuse any arbitrary well-formed JSON data format with no restrictions imposed.  It is recommended, but not\\nrequired, that they follow the same format used for Remote Issue Links, as described at\\n\\u003ca href=\\\"https://developer.atlassian.com/display/JIRADEV/Fields+in+Remote+Issue+Links\\\"\\u003ehttps://developer.atlassian.com/display/JIRADEV/Fields+in+Remote+Issue+Links\\u003c/a\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"versionId\": {\n      \"description\": \"ID of the version.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"globalId\",\n    \"versionId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRemoteVersionLink tool (Status: 200, Content-Type: application/json)
const GetRemoteVersionLinkResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the version exists and the currently authenticated user has permission to view it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **link** (Type: string):\n      - Example: '{\"rel\":\"issue\",\"url\":\"http://www.example.com/jira/rest/api/2/issue/10000\"}'\n  - **name** (Type: string):\n      - Example: 'Issue 10000'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10000'\n"

// NewGetRemoteVersionLinkMCPTool creates the MCP Tool instance for GetRemoteVersionLink
func NewGetRemoteVersionLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRemoteVersionLink",
		"Get specific remote version link - Returns the remote version link associated with the given version ID and global ID.",
		[]byte(GetRemoteVersionLinkInputSchema),
	)
}

// GetRemoteVersionLinkHandler is the handler function for the GetRemoteVersionLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRemoteVersionLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/version/{versionId}/remotelink/{globalId}", args, []string{"globalId", "versionId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetRemoteVersionLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
