package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UpdateUserSettings tool
const UpdateUserSettingsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"User settings update request\",\n      \"properties\": {\n        \"locale\": {\n          \"example\": \"en_GB\",\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"example\": \"user1\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the UpdateUserSettings tool (Status: 400, Content-Type: application/json)
const UpdateUserSettingsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if a error occurs while updating user settings.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateUserSettings tool (Status: 401, Content-Type: application/json)
const UpdateUserSettingsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the user is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateUserSettings tool (Status: 403, Content-Type: application/json)
const UpdateUserSettingsResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the user tries to update settings without appropriate permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateUserSettingsMCPTool creates the MCP Tool instance for UpdateUserSettings
func NewUpdateUserSettingsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateUserSettings",
		"Update a user's preference settings - Update the specified user's settings including their prefered language setting.\n\nValues:\n- Username cannot be blank.\n- The user's locale preference can be removed by setting it to \"None\".\n\nExample PUT request URI(s):\n"+"\x60"+"http://example.com/confluence/rest/api/user/settings"+"\x60"+"\n",
		[]byte(UpdateUserSettingsInputSchema),
	)
}

// UpdateUserSettingsHandler is the handler function for the UpdateUserSettings tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateUserSettingsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/user/settings", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateUserSettings"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
