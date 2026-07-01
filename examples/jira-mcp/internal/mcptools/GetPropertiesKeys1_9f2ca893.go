package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPropertiesKeys1_9f2ca893 tool
const GetPropertiesKeys1_9f2ca893InputSchema = "{\n  \"properties\": {\n    \"commentId\": {\n      \"description\": \"the comment from which keys will be returned.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"commentId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPropertiesKeys1_9f2ca893 tool (Status: 200, Content-Type: application/json)
const GetPropertiesKeys1_9f2ca893ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all properties in the comment.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **keys** (Type: array):\n    - **Items** (Type: object):\n      - **key** (Type: string):\n          - Example: 'issue.support'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/EX-2/properties/issue.support'\n"

// NewGetPropertiesKeys1_9f2ca893MCPTool creates the MCP Tool instance for GetPropertiesKeys1_9f2ca893
func NewGetPropertiesKeys1_9f2ca893MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPropertiesKeys1_9f2ca893",
		"Get properties keys of a comment - Returns the keys of all properties for the comment identified by the key or by the id.",
		[]byte(GetPropertiesKeys1_9f2ca893InputSchema),
	)
}

// GetPropertiesKeys1_9f2ca893Handler is the handler function for the GetPropertiesKeys1_9f2ca893 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPropertiesKeys1_9f2ca893Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/comment/{commentId}/properties", args, []string{"commentId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPropertiesKeys1_9f2ca893")
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
