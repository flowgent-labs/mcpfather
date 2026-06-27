package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Search tool
const SearchInputSchema = "{\n  \"properties\": {\n    \"cql\": {\n      \"description\": \"  a cql query string to use to locate content.\",\n      \"type\": \"string\"\n    },\n    \"cqlcontext\": {\n      \"description\": \" the context to execute a cql search in, this is the json serialized form of SearchContext\",\n      \"type\": \"string\"\n    },\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the content. Default value: \\u003ccode\\u003ehistory,space,version\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"the limit of the number of items to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"the start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"xoauth_requestor_id\": {\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Search tool (Status: 200, Content-Type: application/json)
const SearchResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a paginated list of content.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n"

// Response Template for the Search tool (Status: 404, Content-Type: application/json)
const SearchResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if the CQL is invalid or missing.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewSearchMCPTool creates the MCP Tool instance for Search
func NewSearchMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Search",
		"Search content using CQL - Fetch a list of content using the Confluence Query Language (CQL). See: [Advanced searching using CQL](https://developer.atlassian.com/display/CONFDEV/Advanced+Searching+using+CQL) \n\n Example request URI(s): \n\n- "+"\x60"+"http://localhost:8080/confluence/rest/api/content/search?cql=creator=currentUser()&cqlcontext={\"spaceKey\":\"TST\", \"contentId\":\"55\"}"+"\x60"+"\n- "+"\x60"+"http://localhost:8080/confluence/rest/api/content/search?cql=space=DEV AND label=docs&expand=space,metadata.labels&limit=10"+"\x60"+"",
		[]byte(SearchInputSchema),
	)
}

// SearchHandler is the handler function for the Search tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SearchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/search", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Search"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
