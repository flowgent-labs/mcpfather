package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the ChangePassword1 tool
const ChangePassword1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"password change details\",\n      \"properties\": {\n        \"newPassword\": {\n          \"example\": \"newPassword\",\n          \"type\": \"string\"\n        },\n        \"oldPassword\": {\n          \"example\": \"oldPassword\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the ChangePassword1 tool (Status: 400, Content-Type: application/json)
const ChangePassword1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  returned if any error occurs while changing user password. Refer the validation rules above.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the ChangePassword1 tool (Status: 401, Content-Type: application/json)
const ChangePassword1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> returned if the user is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the ChangePassword1 tool (Status: 403, Content-Type: application/json)
const ChangePassword1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> returned if current password is wrong or if the user has exceeded number of allowed failed login attempts\n.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewChangePassword1MCPTool creates the MCP Tool instance for ChangePassword1
func NewChangePassword1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ChangePassword1",
		"Change password - Change the password for the current user. \n\n Validation Rules: \n\n- New password supplied cannot be null or blank\n\nExample request URI(s):\n\n"+"\x60"+"http://example.com/confluence/rest/api/user/current/password"+"\x60"+"",
		[]byte(ChangePassword1InputSchema),
	)
}

// ChangePassword1Handler is the handler function for the ChangePassword1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ChangePassword1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/user/current/password", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ChangePassword1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
