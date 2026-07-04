package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPropertiesKeys_a29312ee tool
const GetPropertiesKeys_a29312eeInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"description\": \"The id of the board from which property keys will be returned.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"boardId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPropertiesKeys_a29312ee tool (Status: 200, Content-Type: application/json)
const GetPropertiesKeys_a29312eeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested property keys.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **keys** (Type: array):\n    - **Items** (Type: object):\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/EX-2/properties/issue.support'\n      - **key** (Type: string):\n          - Example: 'issue.support'\n"

// NewGetPropertiesKeys_a29312eeMCPTool creates the MCP Tool instance for GetPropertiesKeys_a29312ee
func NewGetPropertiesKeys_a29312eeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPropertiesKeys_a29312ee",
		"Get all properties keys for a board - Returns the keys of all properties for the board identified by the id. The user who retrieves the property keys is required to have permissions to view the board.",
		[]byte(GetPropertiesKeys_a29312eeInputSchema),
	)
}

// GetPropertiesKeys_a29312eeHandler is the handler function for the GetPropertiesKeys_a29312ee tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPropertiesKeys_a29312eeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/properties", args, []string{"boardId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPropertiesKeys_a29312ee")
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
