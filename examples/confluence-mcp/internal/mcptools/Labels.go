package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Labels tool
const LabelsInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 200,\n      \"description\": \"the limit of the number of labels to return, this may be restricted by fixed system limits\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"prefix\": {\n      \"description\": \"the prefixes to filter the labels with.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"start\": {\n      \"description\": \"he start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Labels tool (Status: 200, Content-Type: application/json)
const LabelsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a JSON representation of the existing labels on the content with the given id\n\n## Response Structure\n\n- Structure (Type: object):\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n"

// Response Template for the Labels tool (Status: 404, Content-Type: application/json)
const LabelsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n>  Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewLabelsMCPTool creates the MCP Tool instance for Labels
func NewLabelsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Labels",
		"Get labels - Returns the list of labels on a piece of Content. Example request URI(s): \n\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/label"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/label?prefix=global&start=0&limit=200"+"\x60"+"",
		[]byte(LabelsInputSchema),
	)
}

// LabelsHandler is the handler function for the Labels tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func LabelsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/label", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Labels")
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
