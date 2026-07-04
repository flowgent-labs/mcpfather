package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetActiveUsers tool
const GetActiveUsersInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"properties to expand on the user.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 100,\n      \"description\": \"the limit of the number of users to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"default\": 0,\n      \"description\": \"the start point of the collection to return. This must be non-negative. Default value is 0.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetActiveUsers tool (Status: 200, Content-Type: application/json)
const GetActiveUsersResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a paginated collection of users.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n"

// Response Template for the GetActiveUsers tool (Status: 403, Content-Type: application/json)
const GetActiveUsersResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling user does not have permission to view users. This is possible for anonymous or un-licensed users.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetActiveUsersMCPTool creates the MCP Tool instance for GetActiveUsers
func NewGetActiveUsersMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetActiveUsers",
		"Get active users - Gets a paginated collection of all active users (users which count into license usage).\nThis will exclude users that are:\n- anonymous,\n- deactivated,\n- externally deleted,\n- shadowed\n- unlicensed\n\nThis feature relies on search index and might not be accurate when site reindex is in progress.\n\nDepending on the type of the user the response can include the following fields:\n- "+"\x60"+"email"+"\x60"+": The user's email address.\n- "+"\x60"+"lastLogin"+"\x60"+": The date and time of the user's last successful login. Required \"lastLogin\" expansion.\n- "+"\x60"+"type"+"\x60"+": The type of user (e.g., "+"\x60"+"known"+"\x60"+", "+"\x60"+"anonymous"+"\x60"+").\n- "+"\x60"+"username"+"\x60"+": The user's username.\n- "+"\x60"+"userKey"+"\x60"+": The unique key identifying the user.\n- "+"\x60"+"displayName"+"\x60"+": The user's full display name.\n\nExample request URI(s):\n"+"\x60"+"http://example.com/confluence/rest/api/admin/users/list/active"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/admin/users/list/active?start=0"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/admin/users/list/active?start=0&limit=100"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/admin/users/list/active?start=0&limit=100&expand=status"+"\x60"+"\n",
		[]byte(GetActiveUsersInputSchema),
	)
}

// GetActiveUsersHandler is the handler function for the GetActiveUsers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetActiveUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/admin/users/list/active", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetActiveUsers")
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
