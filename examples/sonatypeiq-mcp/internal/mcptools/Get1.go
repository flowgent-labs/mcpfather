package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the Get1 tool
const Get1InputSchema = "{\n  \"properties\": {\n    \"realm\": {\n      \"default\": \"Internal\",\n      \"description\": \"Enter the " + "\x60" + "realm" + "\x60" + ". Allowed values are " + "\x60" + "Internal" + "\x60" + "," + "\x60" + "OAUTH2" + "\x60" + ", and " + "\x60" + "SAML" + "\x60" + ".\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"Enter the username.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Get1 tool (Status: 200, Content-Type: application/json)
const Get1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains details for the specified user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **username** (Type: string):\n  - **email** (Type: string):\n  - **firstName** (Type: string):\n  - **lastName** (Type: string):\n  - **password** (Type: string):\n  - **realm** (Type: string):\n"

// NewGet1MCPTool creates the MCP Tool instance for Get1
func NewGet1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Get1",
		"Use this method to retrieve user details for the specified user.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(Get1InputSchema),
	)
}

// Get1Handler is the handler function for the Get1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Get1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/users/{username}", args, []string{"username"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Get1")
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
