package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the PolicyCheckCreateUser tool
const PolicyCheckCreateUserInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The intended parameters for the user that would be created.\\nThe username and new password must be specified.  The old password should be specified for\\nupdates where the user would be required to enter it and omitted for those like a password\\nreset or forced change by the administrator where the old password would not be known.\",\n      \"properties\": {\n        \"displayName\": {\n          \"example\": \"Fred Normal\",\n          \"type\": \"string\"\n        },\n        \"emailAddress\": {\n          \"example\": \"fred@example.com\",\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"example\": \"secret\",\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"example\": \"fred\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PolicyCheckCreateUser tool (Status: 200, Content-Type: application/json)
const PolicyCheckCreateUserResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON array of the user-facing messages.\n\n## Response Structure\n\n- Structure (Type: string):\n"

// NewPolicyCheckCreateUserMCPTool creates the MCP Tool instance for PolicyCheckCreateUser
func NewPolicyCheckCreateUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PolicyCheckCreateUser",
		"Get reasons for password policy disallowance on user creation - Returns a list of statements explaining why the password policy would disallow a proposed password for a new user.\nYou can use this method to test the password policy validation. This could be done prior to an action\nwhere a new user and related password are created, using methods like the ones in\n<a href=\"https://docs.atlassian.com/jira/latest/com/atlassian/jira/bc/user/UserService.html\">UserService</a>.\nFor example, you could use this to validate a password in a create user form in the user interface, as the user enters it.\nThe username and new password must be not empty to perform the validation.\nNote, this method will help you validate against the policy only. It won't check any other validations that might be performed\nwhen creating a new user, e.g. checking whether a user with the same name already exists.\n",
		[]byte(PolicyCheckCreateUserInputSchema),
	)
}

// PolicyCheckCreateUserHandler is the handler function for the PolicyCheckCreateUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PolicyCheckCreateUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/password/policy/createUser", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "PolicyCheckCreateUser"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
