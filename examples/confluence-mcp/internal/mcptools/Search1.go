package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Search1 tool
const Search1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"type\": \"string\"\n    },\n    \"cql\": {\n      \"description\": \"the CQL query see \\u003ca href='https://developer.atlassian.com/confdev/confluence-rest-api/advanced-searching-using-cql'\\u003eadvanced searching in confluence using CQL\\u003c/a\\u003e\",\n      \"type\": \"string\"\n    },\n    \"cqlcontext\": {\n      \"description\": \"the execution context for CQL functions, provides current space key and content id. If this is not provided some CQL functions will not be available.\",\n      \"type\": \"string\"\n    },\n    \"excerpt\": {\n      \"description\": \"the excerpt strategy to apply to the result, one of : indexed, highlight, none. This defaults to highlight.\",\n      \"type\": \"string\"\n    },\n    \"expand\": {\n      \"description\": \"the properties to expand on the search result, this may cause database requests for some properties\",\n      \"type\": \"string\"\n    },\n    \"includeArchivedSpaces\": {\n      \"default\": false,\n      \"description\": \"whether to include content in archived spaces in the result, this defaults to false.\",\n      \"type\": \"boolean\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"the limit of the number of items to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"default\": 0,\n      \"description\": \"he start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Search1 tool (Status: 200, Content-Type: application/json)
const Search1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a full JSON representation of a list of search results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n"

// Response Template for the Search1 tool (Status: 400, Content-Type: application/json)
const Search1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if the query cannot be parsed\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewSearch1MCPTool creates the MCP Tool instance for Search1
func NewSearch1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Search1",
		"Search for entities in confluence - Search for entities in Confluence using the [Confluence Query Language (CQL)](https://developer.atlassian.com/confdev/confluence-rest-api/advanced-searching-using-cql). For example:\n\nExample request URI(s):\n\n- "+"\x60"+"http://localhost:8080/confluence/rest/api/search?cql=creator=currentUser()&type%20in%20(space,page,user)&cqlcontext={\"spaceKey\":\"TST\", \"contentId\":\"55\"}"+"\x60"+"\n\n- "+"\x60"+"http://localhost:8080/confluence/rest/api/search?cql=siteSearch~'example'%20AND%20label=docs&expand=content.space,space.homepage&limit=10"+"\x60"+"",
		[]byte(Search1InputSchema),
	)
}

// Search1Handler is the handler function for the Search1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Search1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/search", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Search1")
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
