package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPropertyKeys tool
const GetPropertyKeysInputSchema = "{\n  \"properties\": {\n    \"issueTypeId\": {\n      \"description\": \"The issue type from which the keys will be returned.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueTypeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPropertyKeys tool (Status: 200, Content-Type: application/json)
const GetPropertyKeysResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns keys of all properties for the issue type.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **keys** (Type: array):\n    - **Items** (Type: object):\n      - **key** (Type: string):\n          - Example: 'issue.support'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/EX-2/properties/issue.support'\n"

// NewGetPropertyKeysMCPTool creates the MCP Tool instance for GetPropertyKeys
func NewGetPropertyKeysMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPropertyKeys",
		"Get all properties keys for issue type - Returns the keys of all properties for the issue type identified by the id",
		[]byte(GetPropertyKeysInputSchema),
	)
}

// GetPropertyKeysHandler is the handler function for the GetPropertyKeys tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPropertyKeysHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuetype/{issueTypeId}/properties", args, []string{"issueTypeId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPropertyKeys")
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
