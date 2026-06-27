package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPasswordPolicy tool
const GetPasswordPolicyInputSchema = "{\n  \"properties\": {\n    \"hasOldPassword\": {\n      \"default\": false,\n      \"description\": \"Whether or not the user will be required to enter their current password.  Use false (the default) if this is a new user or if an administrator is forcibly changing another user's password.\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPasswordPolicy tool (Status: 200, Content-Type: application/json)
const GetPasswordPolicyResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON array of the user-facing messages.\n\n## Response Structure\n\n- Structure (Type: string):\n"

// NewGetPasswordPolicyMCPTool creates the MCP Tool instance for GetPasswordPolicy
func NewGetPasswordPolicyMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPasswordPolicy",
		"Get current password policy requirements - Returns the list of requirements for the current password policy. For example, \"The password must have at least 10 characters.\", \"The password must not be similar to the user's name or email address.\", etc.",
		[]byte(GetPasswordPolicyInputSchema),
	)
}

// GetPasswordPolicyHandler is the handler function for the GetPasswordPolicy tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPasswordPolicyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/password/policy", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPasswordPolicy"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
