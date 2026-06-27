package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the PolicyCheckUpdateUser tool
const PolicyCheckUpdateUserInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The intended parameters for the update that would be performed.\\nThe username and new password must be specified.  The old password should be specified for\\nupdates where the user would be required to enter it and omitted for those like a password\\nreset or forced change by the administrator where the old password would not be known.\",\n      \"properties\": {\n        \"newPassword\": {\n          \"example\": \"correcthorsebatterystaple\",\n          \"type\": \"string\"\n        },\n        \"oldPassword\": {\n          \"example\": \"secret\",\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"example\": \"fred\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PolicyCheckUpdateUser tool (Status: 200, Content-Type: application/json)
const PolicyCheckUpdateUserResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON array of the user-facing messages. If no policy is set, then his will be an empty list.\n\n## Response Structure\n\n- Structure (Type: string):\n"

// NewPolicyCheckUpdateUserMCPTool creates the MCP Tool instance for PolicyCheckUpdateUser
func NewPolicyCheckUpdateUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PolicyCheckUpdateUser",
		"Get reasons for password policy disallowance on user password update - Returns a list of statements explaining why the password policy would disallow a proposed new password for a user with an existing password.\nYou can use this method to test the password policy validation. This could be done prior to an action where the password\nis actually updated, using methods like ChangePassword or ResetPassword.\nFor example, you could use this to validate a password in a change password form in the user interface, as the user enters it.\nThe user must exist and the username and new password must be not empty, to perform the validation.\nNote, this method will help you validate against the policy only. It won't check any other validations that might be performed\nwhen submitting a password change/reset request, e.g. verifying whether the old password is valid.\n",
		[]byte(PolicyCheckUpdateUserInputSchema),
	)
}

// PolicyCheckUpdateUserHandler is the handler function for the PolicyCheckUpdateUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PolicyCheckUpdateUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/password/policy/updateUser", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "PolicyCheckUpdateUser"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
