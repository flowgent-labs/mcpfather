package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetUsers tool
const GetUsersInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"properties to expand on the user.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 100,\n      \"description\": \"the limit of the number of users to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"default\": 0,\n      \"description\": \"the start point of the collection to return. This must be non-negative. Default value is 0.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetUsers tool (Status: 200, Content-Type: application/json)
const GetUsersResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a paginated collection of users.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **_links** (Type: object):\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n"

// Response Template for the GetUsers tool (Status: 403, Content-Type: application/json)
const GetUsersResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling user does not have permission to view users. This is possible for anonymous or un-licensed users.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetUsersMCPTool creates the MCP Tool instance for GetUsers
func NewGetUsersMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUsers",
		"Get registered users - Gets a paginated collection of all registered users, including but not limited to:\n\n- Disabled users\n- Enabled users\n- Enabled users which count towards the license count on the site\n- Enabled users which do not count towards the license count on the site\n- Enabled users which have \"can use\" global permissions\n- Enabled users which do not have \"can use\" global permissions\n\nExample request URI(s):\n\n"+"\x60"+"http://example.com/confluence/rest/api/user/list"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/list?start=0"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/list?start=0&limit=100"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/list?start=0&limit=100&expand=status"+"\x60"+"",
		[]byte(GetUsersInputSchema),
	)
}

// GetUsersHandler is the handler function for the GetUsers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/user/list", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetUsers")
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
