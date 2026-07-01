package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetGroups1 tool
const GetGroups1InputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"properties to expand on the user.\",\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"userkey of the user to request from this resource\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 200,\n      \"description\": \"the limit of the number of users to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"the start point of the collection to return. This must be non-negative. Default value is 0.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"username\": {\n      \"description\": \"userName of the user to get the groups for.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetGroups1 tool (Status: 200, Content-Type: application/json)
const GetGroups1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of a user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n"

// Response Template for the GetGroups1 tool (Status: 403, Content-Type: application/json)
const GetGroups1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling user does not have permission to use confluence.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetGroups1MCPTool creates the MCP Tool instance for GetGroups1
func NewGetGroups1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetGroups1",
		"Get groups - Get a paginated collection of groups that the given user is a member of. Example request URI(s):\n\n"+"\x60"+"http://example.com/confluence/rest/api/user/memberof?username=jblogs"+"\x60"+"\n"+"\x60"+"http://example.com/confluence/rest/api/user/memberof?key=402880824ff933a4014ff9345d7c0002"+"\x60"+"",
		[]byte(GetGroups1InputSchema),
	)
}

// GetGroups1Handler is the handler function for the GetGroups1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetGroups1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/user/memberof", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetGroups1")
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
