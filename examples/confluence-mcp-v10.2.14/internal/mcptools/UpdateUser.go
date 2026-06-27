package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UpdateUser tool
const UpdateUserInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Details of the user to be updated\",\n      \"properties\": {\n        \"currentPassword\": {\n          \"example\": \"password\",\n          \"type\": \"string\"\n        },\n        \"email\": {\n          \"example\": \"someuser@someemail.com\",\n          \"type\": \"string\"\n        },\n        \"fullName\": {\n          \"example\": \"Some User\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"username\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateUser tool (Status: 204, Content-Type: application/json)
const UpdateUserResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 204\n\n**Content-Type:** application/json\n\n> Returned if the update was successful.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateUser tool (Status: 400, Content-Type: application/json)
const UpdateUserResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if any error occurs while updating the user\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateUser tool (Status: 401, Content-Type: application/json)
const UpdateUserResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> returned if an anonymous (or unauthenticated) user tries to update a user\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateUser tool (Status: 403, Content-Type: application/json)
const UpdateUserResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> returned if user does not have enough permission to update a user\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateUserMCPTool creates the MCP Tool instance for UpdateUser
func NewUpdateUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateUser",
		"Update user - Updates the user identified by the username. The following fields can be updated: email, full name.\n\"**Requirements**:\n- The fullName should not be blank\n- The fullName should not contain any forbidden characters (< or >)\n- The fullName should not be anonymous (in english or other system locale)\n- The email should not be blank\n- The email should be a valid email address\n",
		[]byte(UpdateUserInputSchema),
	)
}

// UpdateUserHandler is the handler function for the UpdateUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/admin/user/{username}", args, []string{"username"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateUser"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
