package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetUserList tool
const GetUserListInputSchema = "{\n  \"properties\": {\n    \"cursor\": {\n      \"description\": \"The position in the stream to continue iterating over all users.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"maxResults\": {\n      \"default\": 2000,\n      \"description\": \"The maximum number of users to return per page (defaults to 2000).\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetUserList tool (Status: 200, Content-Type: application/json)
const GetUserListResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of users.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **maxResults** (Type: integer, int32):\n  - **nextCursor** (Type: string):\n  - **nextPage** (Type: string, uri):\n  - **self** (Type: string, uri):\n  - **values** (Type: array):\n    - **Items** (Type: object):\n  - **isLast** (Type: boolean):\n"

// NewGetUserListMCPTool creates the MCP Tool instance for GetUserList
func NewGetUserListMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUserList",
		"List all users - Returns a list of all users. This resource cannot be accessed anonymously.\nThis Api is a streaming-like endpoint. For performance and security  reasons, it is not indicating the total\nnumber of users available in the system. The first call should be done without the cursor parameter.\nSubsequent calls should use the value of the next cursor returned in the previous call. Specific values of\ncursor are not guaranteed to be valid in the future and are not part of the API, so they should not be used\nas a key for caching or storing data. The order in which the users are returned is not defined. It is guaranteed\nthat the same user will not be returned twice in the sequence of calls. For resiliency reason this endpoint\nnever returns 404 code, even if called with a cursor parameter that was not returned in the previous call.\n",
		[]byte(GetUserListInputSchema),
	)
}

// GetUserListHandler is the handler function for the GetUserList tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUserListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/list", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetUserList")
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
