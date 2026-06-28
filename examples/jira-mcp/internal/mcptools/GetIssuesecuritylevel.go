package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetIssuesecuritylevel tool
const GetIssuesecuritylevelInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"An issue security level id\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssuesecuritylevel tool (Status: 200, Content-Type: application/json)
const GetIssuesecuritylevelResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the issue type exists and is visible by the calling user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n      - Example: 'This is a security level'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **name** (Type: string):\n      - Example: 'My Security Level'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/securitylevel/10000'\n"

// NewGetIssuesecuritylevelMCPTool creates the MCP Tool instance for GetIssuesecuritylevel
func NewGetIssuesecuritylevelMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssuesecuritylevel",
		"Get a security level by ID - Returns a full representation of the security level that has the given id.",
		[]byte(GetIssuesecuritylevelInputSchema),
	)
}

// GetIssuesecuritylevelHandler is the handler function for the GetIssuesecuritylevel tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssuesecuritylevelHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/securitylevel/{id}", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetIssuesecuritylevel")
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
