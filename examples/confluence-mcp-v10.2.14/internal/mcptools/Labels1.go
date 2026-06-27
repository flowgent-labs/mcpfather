package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Labels1 tool
const Labels1InputSchema = "{\n  \"properties\": {\n    \"labelName\": {\n      \"description\": \"The name of the label (excluding any prefix)\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 100,\n      \"description\": \"The limit of the number of labels to return, this may be restricted by fixed system limit.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"namespace\": {\n      \"description\": \"The namespace of the labels.\",\n      \"type\": \"string\"\n    },\n    \"owner\": {\n      \"description\": \"The owner of the labels.\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"The spaceKey to restrict by.\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"The start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Labels1 tool (Status: 200, Content-Type: application/json)
const Labels1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Return a list of labels matching the given request.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n  - **limit** (Type: number):\n      - Example: '25'\n"

// Response Template for the Labels1 tool (Status: 400, Content-Type: application/json)
const Labels1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Return a bad request error if the given label name is invalid\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Labels1 tool (Status: 404, Content-Type: application/json)
const Labels1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Return a not found error if the given label name is not found\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewLabels1MCPTool creates the MCP Tool instance for Labels1
func NewLabels1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Labels1",
		"Get list of labels matching the given label name, namespace, space (via space key) or owner. - Returns a paginated list of labels matching the given label name, namespace, space (via space key) or owner.\nLeave query params empty to ignore.\n\nExample request URI(s):\n"+"\x60"+"http://example.com/confluence/rest/api/label/labels?spaceKey=MYS&namespace=global&limit=3"+"\x60"+"",
		[]byte(Labels1InputSchema),
	)
}

// Labels1Handler is the handler function for the Labels1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Labels1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/label/labels", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Labels1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
