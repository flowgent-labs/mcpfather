package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAllFields tool
const GetAllFieldsInputSchema = "{\n  \"properties\": {\n    \"projectKey\": {\n      \"description\": \"the key of the project; this parameter is optional\",\n      \"type\": \"string\"\n    },\n    \"screenId\": {\n      \"description\": \"id of screen\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"tabId\": {\n      \"description\": \"id of tab\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"screenId\",\n    \"tabId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAllFields tool (Status: 200, Content-Type: application/json)
const GetAllFieldsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all fields for the given tab.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **showWhenEmpty** (Type: boolean):\n      - Example: 'false'\n  - **type** (Type: string):\n      - Example: 'The type of the field. One of: 'system', 'custom', 'jira'.'\n  - **id** (Type: string):\n      - Example: 'summary'\n  - **name** (Type: string):\n      - Example: 'Summary'\n"

// NewGetAllFieldsMCPTool creates the MCP Tool instance for GetAllFields
func NewGetAllFieldsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllFields",
		"Get all fields for a tab - Gets all fields for a given tab.",
		[]byte(GetAllFieldsInputSchema),
	)
}

// GetAllFieldsHandler is the handler function for the GetAllFields tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllFieldsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/screens/{screenId}/tabs/{tabId}/fields", args, []string{"screenId", "tabId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAllFields")
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
