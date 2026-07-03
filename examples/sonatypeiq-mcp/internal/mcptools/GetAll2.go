package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAll2 tool
const GetAll2InputSchema = "{\n  \"properties\": {\n    \"realm\": {\n      \"default\": \"Internal\",\n      \"description\": \"Enter the " + "\x60" + "realm" + "\x60" + ". Allowed values are " + "\x60" + "Internal" + "\x60" + "," + "\x60" + "OAUTH2" + "\x60" + ", and " + "\x60" + "SAML" + "\x60" + ".\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetAll2 tool (Status: 200, Content-Type: application/json)
const GetAll2ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains user details. Passwords are excluded for security.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **users** (Type: array):\n    - **Items** (Type: object):\n      - **username** (Type: string):\n      - **email** (Type: string):\n      - **firstName** (Type: string):\n      - **lastName** (Type: string):\n      - **password** (Type: string):\n      - **realm** (Type: string):\n"

// NewGetAll2MCPTool creates the MCP Tool instance for GetAll2
func NewGetAll2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAll2",
		"Use this method to retrieve user details for all users.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetAll2InputSchema),
	)
}

// GetAll2Handler is the handler function for the GetAll2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAll2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/users", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAll2")
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
