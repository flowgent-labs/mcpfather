package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the RevokePermissionsFromAnonymousUser tool
const RevokePermissionsFromAnonymousUserInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"items\": {\n        \"properties\": {\n          \"operationKey\": {\n            \"example\": \"read\",\n            \"type\": \"string\"\n          },\n          \"targetType\": {\n            \"example\": \"space\",\n            \"type\": \"string\"\n          }\n        },\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    },\n    \"spaceKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the RevokePermissionsFromAnonymousUser tool (Status: 400, Content-Type: application/json)
const RevokePermissionsFromAnonymousUserResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if incorrect permissions are passed in request (for e.g. non existing operation or global permission).\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the RevokePermissionsFromAnonymousUser tool (Status: 401, Content-Type: application/json)
const RevokePermissionsFromAnonymousUserResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the RevokePermissionsFromAnonymousUser tool (Status: 403, Content-Type: application/json)
const RevokePermissionsFromAnonymousUserResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewRevokePermissionsFromAnonymousUserMCPTool creates the MCP Tool instance for RevokePermissionsFromAnonymousUser
func NewRevokePermissionsFromAnonymousUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RevokePermissionsFromAnonymousUser",
		"Revoke space permissions from anonymous user - Revoke permissions from anonymous user in the given space.\nIf anonymous user doesn't have permissions that we are trying to revoke, those permissions will be silently skipped.\nMultiple permissions could be passed in one request. Supported targetType and operationKey pairs:\n* space read\n* space administer\n* space export\n* space restrict\n* space delete_own\n* space delete_mail\n* page create\n* page delete\n* blogpost create\n* blogpost delete\n* comment create\n* comment delete\n* attachment create\n* attachment delete\n\nSee <a href=\"https://confluence.atlassian.com/display/DOC/Space+Permissions+Overview\">Space Permissions documentation</a> for additional information about supported permissions.\n\nExample request URI's:\n"+"\x60"+"https://example.com/confluence/rest/api/space/TESTSPACE/permissions/anonymous/revoke"+"\x60"+"",
		[]byte(RevokePermissionsFromAnonymousUserInputSchema),
	)
}

// RevokePermissionsFromAnonymousUserHandler is the handler function for the RevokePermissionsFromAnonymousUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RevokePermissionsFromAnonymousUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/space/{spaceKey}/permissions/anonymous/revoke", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RevokePermissionsFromAnonymousUser"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
