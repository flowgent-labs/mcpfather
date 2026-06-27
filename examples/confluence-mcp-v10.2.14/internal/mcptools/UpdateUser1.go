package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UpdateUser1 tool
const UpdateUser1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"User details\",\n      \"properties\": {\n        \"currentPassword\": {\n          \"example\": \"password\",\n          \"type\": \"string\"\n        },\n        \"email\": {\n          \"example\": \"someuser@someemail.com\",\n          \"type\": \"string\"\n        },\n        \"fullName\": {\n          \"example\": \"Some User\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the UpdateUser1 tool (Status: 400, Content-Type: application/json)
const UpdateUser1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  Returned if any error occurs while updating user details. Refer the validation rules above.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateUser1 tool (Status: 401, Content-Type: application/json)
const UpdateUser1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the user is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateUser1 tool (Status: 403, Content-Type: application/json)
const UpdateUser1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if current password is wrong or if the user has exceeded number of allowed failed login attempts\n.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateUser1MCPTool creates the MCP Tool instance for UpdateUser1
func NewUpdateUser1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateUser1",
		"Update details of the current user - Change the current user's details.\n\nValidation Rules:\n- Full name cannot be blank, containing <> characters or be reserved by Confluence.\n- Email must be a valid email address.\n- Current password must be supplied for changing email address.\n\nExample PUT request URI(s):\n"+"\x60"+"http://example.com/confluence/rest/api/user/current"+"\x60"+"\n",
		[]byte(UpdateUser1InputSchema),
	)
}

// UpdateUser1Handler is the handler function for the UpdateUser1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateUser1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/user/current", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateUser1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
