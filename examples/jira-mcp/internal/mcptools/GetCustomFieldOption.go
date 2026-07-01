package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetCustomFieldOption tool
const GetCustomFieldOptionInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"a String containing an Custom Field Option id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetCustomFieldOption tool (Status: 200, Content-Type: application/json)
const GetCustomFieldOptionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the Custom Field Option exists and is visible by the calling user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **disabled** (Type: boolean):\n      - Example: 'false'\n  - **id** (Type: integer, int64):\n      - Example: '3'\n  - **self** (Type: string, uri):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/customFieldOption/3'\n  - **value** (Type: string):\n      - Example: 'Blue'\n  - **childrenIds** (Type: array):\n      - Example: '[4,5]'\n    - **Items** (Type: integer, int64):\n"

// NewGetCustomFieldOptionMCPTool creates the MCP Tool instance for GetCustomFieldOption
func NewGetCustomFieldOptionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCustomFieldOption",
		"Get custom field option by ID - Returns a full representation of the Custom Field Option that has the given id.",
		[]byte(GetCustomFieldOptionInputSchema),
	)
}

// GetCustomFieldOptionHandler is the handler function for the GetCustomFieldOption tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCustomFieldOptionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/customFieldOption/{id}", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetCustomFieldOption")
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
