package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the Get4 tool
const Get4InputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"the key of the role to use.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Get4 tool (Status: 200, Content-Type: application/json)
const Get4ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the ApplicationRole if it exists.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **userCount** (Type: integer, int32):\n      - Example: '5'\n  - **key** (Type: string):\n      - Example: 'jira-software'\n  - **numberOfSeats** (Type: integer, int32):\n      - Example: '10'\n  - **userCountDescription** (Type: string):\n      - Example: '5 developers'\n  - **hasUnlimitedSeats** (Type: boolean):\n      - Example: 'false'\n  - **name** (Type: string):\n      - Example: 'Jira Software'\n  - **platform** (Type: boolean):\n      - Example: 'false'\n  - **selectedByDefault** (Type: boolean):\n      - Example: 'false'\n  - **remainingSeats** (Type: integer, int32):\n      - Example: '5'\n  - **defaultGroups** (Type: array):\n      - Unique Items: true\n      - Example: '[\"jira-software-users\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-software-users\"]'\n  - **defined** (Type: boolean):\n      - Example: 'false'\n  - **groups** (Type: array):\n      - Unique Items: true\n      - Example: '[\"jira-software-users\",\"jira-testers\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-software-users\",\"jira-testers\"]'\n"

// NewGet4MCPTool creates the MCP Tool instance for Get4
func NewGet4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Get4",
		"Get application role by key - Returns the ApplicationRole with passed key if it exists.",
		[]byte(Get4InputSchema),
	)
}

// Get4Handler is the handler function for the Get4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Get4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/applicationrole/{key}", args, []string{"key"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Get4")
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
