package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetRemoteVersionLinks tool
const GetRemoteVersionLinksInputSchema = "{\n  \"properties\": {\n    \"globalId\": {\n      \"description\": \"The id of the remote issue link to be returned.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetRemoteVersionLinks tool (Status: 200, Content-Type: application/json)
const GetRemoteVersionLinksResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the remote version links are successfully retrieved.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **links** (Type: array):\n    - **Items** (Type: object):\n      - **name** (Type: string):\n          - Example: 'Issue 10000'\n      - **self** (Type: string, uri):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/10000'\n      - **link** (Type: string):\n          - Example: '{\"rel\":\"issue\",\"url\":\"http://www.example.com/jira/rest/api/2/issue/10000\"}'\n"

// NewGetRemoteVersionLinksMCPTool creates the MCP Tool instance for GetRemoteVersionLinks
func NewGetRemoteVersionLinksMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRemoteVersionLinks",
		"Get remote version links by global ID - Returns the remote version links for a given global ID.",
		[]byte(GetRemoteVersionLinksInputSchema),
	)
}

// GetRemoteVersionLinksHandler is the handler function for the GetRemoteVersionLinks tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRemoteVersionLinksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/version/remotelink", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRemoteVersionLinks")
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
