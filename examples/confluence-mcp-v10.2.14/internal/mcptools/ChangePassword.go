package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the ChangePassword tool
const ChangePasswordInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"password\": {\n          \"example\": \"password\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"username\": {\n      \"description\": \"the username identifying the given user\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ChangePassword tool (Status: 400, Content-Type: application/json)
const ChangePasswordResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if any error occurs while changing user password. Refer to the validation rules above.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the ChangePassword tool (Status: 403, Content-Type: application/json)
const ChangePasswordResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> returned if user does not have enough permission to change another user's password. User should be a System admin\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the ChangePassword tool (Status: 404, Content-Type: application/json)
const ChangePasswordResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> returned if user with specified userName not found\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewChangePasswordMCPTool creates the MCP Tool instance for ChangePassword
func NewChangePasswordMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ChangePassword",
		"Change password - Change the password for the user identified by the username. \n\n**Validation rules** : \n\n- The new password should not be null or blank. \n\n",
		[]byte(ChangePasswordInputSchema),
	)
}

// ChangePasswordHandler is the handler function for the ChangePassword tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ChangePasswordHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/admin/user/{username}/password", args, []string{"username"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ChangePassword"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
