package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GrantPermissionsToGroup tool
const GrantPermissionsToGroupInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"items\": {\n        \"properties\": {\n          \"operationKey\": {\n            \"example\": \"read\",\n            \"type\": \"string\"\n          },\n          \"targetType\": {\n            \"example\": \"space\",\n            \"type\": \"string\"\n          }\n        },\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    },\n    \"groupName\": {\n      \"description\": \"the name of the group to look up.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupName\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GrantPermissionsToGroup tool (Status: 400, Content-Type: application/json)
const GrantPermissionsToGroupResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if incorrect permissions are passed in request (for e.g. non existing operation or global permission).\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GrantPermissionsToGroup tool (Status: 401, Content-Type: application/json)
const GrantPermissionsToGroupResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GrantPermissionsToGroup tool (Status: 403, Content-Type: application/json)
const GrantPermissionsToGroupResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GrantPermissionsToGroup tool (Status: 404, Content-Type: application/json)
const GrantPermissionsToGroupResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if group with specified name not found.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGrantPermissionsToGroupMCPTool creates the MCP Tool instance for GrantPermissionsToGroup
func NewGrantPermissionsToGroupMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GrantPermissionsToGroup",
		"Grants global permissions to a group - Grant global permissions to a group.\n\nOperation doesn't override existing permissions, will only add those one that weren't granted before.\n\nMultiple permissions could be passed in one request. Supported targetType and operationKey pairs:\n\n* application use\n* application administer\n* system administer\n* personal_space create\n* space create\n\nSee <a href=\"https://confluence.atlassian.com/display/DOC/Global+Permissions+Overview\">Global Permissions documentation</a> for additional information about supported permissions.\n\n\nExample request URI's:\n\n"+"\x60"+"https://example.com/confluence/rest/api/permissions/group/test-group-name/grant"+"\x60"+"",
		[]byte(GrantPermissionsToGroupInputSchema),
	)
}

// GrantPermissionsToGroupHandler is the handler function for the GrantPermissionsToGroup tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GrantPermissionsToGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/permissions/group/{groupName}/grant", args, []string{"groupName"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GrantPermissionsToGroup"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
