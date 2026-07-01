package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetIssueAllTypes tool
const GetIssueAllTypesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueAllTypes tool (Status: 200, Content-Type: application/json)
const GetIssueAllTypesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of issue types.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **subtask** (Type: boolean):\n      - Example: 'false'\n  - **avatarId** (Type: integer, int64):\n      - Example: '10002'\n  - **description** (Type: string):\n      - Example: 'A problem which impairs or prevents the functions of the product.'\n  - **iconUrl** (Type: string):\n      - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n  - **id** (Type: string):\n      - Example: '1'\n  - **name** (Type: string):\n      - Example: 'Bug'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n"

// NewGetIssueAllTypesMCPTool creates the MCP Tool instance for GetIssueAllTypes
func NewGetIssueAllTypesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueAllTypes",
		"Get list of all issue types visible to user - Returns a list of all issue types visible to the user",
		[]byte(GetIssueAllTypesInputSchema),
	)
}

// GetIssueAllTypesHandler is the handler function for the GetIssueAllTypes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueAllTypesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuetype", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetIssueAllTypes")
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
