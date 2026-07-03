package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the ScanContent tool
const ScanContentInputSchema = "{\n  \"properties\": {\n    \"cursor\": {\n      \"description\": \"the identifier which is used to skip results from a previous query when paginating. Cursor is empty in first request, to move forward or backward use cursor provided in response.\",\n      \"type\": \"string\"\n    },\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the content. Default value: \\u003ccode\\u003ehistory,space,version\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"the limit of the number of items to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"spaceKey\": {\n      \"description\": \" the space key to find content under.\",\n      \"type\": \"string\"\n    },\n    \"status\": {\n      \"description\": \" list of statuses the content to be found is in. Defaults to current is not specified. If set to 'any', content in 'current' and 'trashed' status will be fetched. Does not support 'historical' status for now.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"type\": {\n      \"description\": \"the content type to return. Default value: \\u003ccode\\u003epage\\u003c/code\\u003e. Valid values: \\u003ccode\\u003epage, blogpost, comment\\u003c/code\\u003e. All types are returned if fetching via list of IDS. Type is only required for first request, latest request will use cursor\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the ScanContent tool (Status: 200, Content-Type: application/json)
const ScanContentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a JSON representation of the list of content.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **size** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/confluence/rest/api/latest/..?limit=25&cursor=cursortype%3Atrue%3A360423'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/confluence/rest/api/..?limit=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/confluence/rest/api/latest/..?limit=25&cursor=cursortype%3Afalse%3A360456'\n  - **cursor** (Type: string):\n      - Example: 'cursortype:false:360422'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **nextCursor** (Type: string):\n      - Example: 'cursortype:false:360456'\n  - **prevCursor** (Type: string):\n      - Example: 'cursortype:true:360423'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n"

// Response Template for the ScanContent tool (Status: 400, Content-Type: application/json)
const ScanContentResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if the cursor is invalid.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the ScanContent tool (Status: 404, Content-Type: application/json)
const ScanContentResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> returned if the user is not permitted.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewScanContentMCPTool creates the MCP Tool instance for ScanContent
func NewScanContentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ScanContent",
		"Scan content by space key - Returns a paginated list of Content. Example request URI(s): \n\n- "+"\x60"+"http://example.com/confluence/rest/api/content/scan?spaceKey=TST&limit=100&expand=space,body.view,version,container"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/scan?limit=100&expand=space,body.view,version,container"+"\x60"+"",
		[]byte(ScanContentInputSchema),
	)
}

// ScanContentHandler is the handler function for the ScanContent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ScanContentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/scan", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ScanContent")
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
