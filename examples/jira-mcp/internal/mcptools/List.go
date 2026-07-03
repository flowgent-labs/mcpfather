package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the List tool
const ListInputSchema = "{\n  \"properties\": {\n    \"filter\": {\n      \"description\": \"An optional filter that is applied to the list of dashboards.\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"A hint as to the maximum number of dashboards to return in each call.\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"description\": \"The index of the first dashboard to return (0-based).\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the List tool (Status: 200, Content-Type: application/json)
const ListResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of dashboards.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **prev** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/dashboard?startAt=0'\n  - **startAt** (Type: integer, int32):\n      - Example: '10'\n  - **total** (Type: integer, int32):\n      - Example: '143'\n  - **dashboards** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n          - Example: '10000'\n      - **name** (Type: string):\n          - Example: 'System Dashboard'\n      - **self** (Type: string):\n          - Example: 'http://localhost:8090/jira/rest/api/2.0/dashboard/10000'\n      - **view** (Type: string):\n          - Example: 'http://localhost:8090/jira/secure/Dashboard.jspa?selectPageId=10000'\n  - **maxResults** (Type: integer, int32):\n      - Example: '10'\n  - **next** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/dashboard?startAt=10'\n"

// NewListMCPTool creates the MCP Tool instance for List
func NewListMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"List",
		"Get all dashboards with optional filtering - Returns a list of all dashboards, optionally filtering them.",
		[]byte(ListInputSchema),
	)
}

// ListHandler is the handler function for the List tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/dashboard", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "List")
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
