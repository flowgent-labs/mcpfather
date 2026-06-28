package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetStatusCategories tool
const GetStatusCategoriesInputSchema = "{\n  \"properties\": {\n    \"request\": {\n      \"description\": \"a Request\",\n      \"type\": \"string\"\n    },\n    \"uriInfo\": {\n      \"description\": \"a UriInfo\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetStatusCategories tool (Status: 200, Content-Type: application/json)
const GetStatusCategoriesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all Jira issue status categories in JSON format, that are visible to the user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **key** (Type: string):\n      - Example: 'new'\n  - **name** (Type: string):\n      - Example: 'To Do'\n  - **self** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/statuscategory/1'\n  - **colorName** (Type: string):\n      - Example: 'blue-gray'\n  - **id** (Type: integer, int64):\n      - Example: '1'\n"

// NewGetStatusCategoriesMCPTool creates the MCP Tool instance for GetStatusCategories
func NewGetStatusCategoriesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetStatusCategories",
		"Get all status categories - Returns a list of all status categories",
		[]byte(GetStatusCategoriesInputSchema),
	)
}

// GetStatusCategoriesHandler is the handler function for the GetStatusCategories tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetStatusCategoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/statuscategory", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetStatusCategories")
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
