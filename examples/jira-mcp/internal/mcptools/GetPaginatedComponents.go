package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPaginatedComponents tool
const GetPaginatedComponentsInputSchema = "{\n  \"properties\": {\n    \"maxResults\": {\n      \"default\": \"100\",\n      \"description\": \"the maximum number of components to return\",\n      \"type\": \"string\"\n    },\n    \"projectIds\": {\n      \"description\": \"the set of project ids to filter components\",\n      \"type\": \"string\"\n    },\n    \"query\": {\n      \"description\": \"the string that components names will be matched with\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"default\": \"0\",\n      \"description\": \"the index of the first components to return\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPaginatedComponents tool (Status: 200, Content-Type: application/json)
const GetPaginatedComponentsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns paginated list of components\n\n## Response Structure\n\n- Structure (Type: object):\n  - **maxResults** (Type: integer, int32):\n  - **nextPage** (Type: string, uri):\n  - **self** (Type: string, uri):\n  - **startAt** (Type: integer, int64):\n  - **total** (Type: integer, int64):\n  - **values** (Type: array):\n    - **Items** (Type: object):\n  - **isLast** (Type: boolean):\n"

// NewGetPaginatedComponentsMCPTool creates the MCP Tool instance for GetPaginatedComponents
func NewGetPaginatedComponentsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPaginatedComponents",
		"Get paginated components - Returns paginated list of filtered active components",
		[]byte(GetPaginatedComponentsInputSchema),
	)
}

// GetPaginatedComponentsHandler is the handler function for the GetPaginatedComponents tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPaginatedComponentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/component/page", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPaginatedComponents")
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
