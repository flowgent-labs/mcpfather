package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the RevokePermissionsFromUser1 tool
const RevokePermissionsFromUser1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"items\": {\n        \"properties\": {\n          \"operationKey\": {\n            \"example\": \"read\",\n            \"type\": \"string\"\n          },\n          \"targetType\": {\n            \"example\": \"space\",\n            \"type\": \"string\"\n          }\n        },\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    },\n    \"spaceKey\": {\n      \"type\": \"string\"\n    },\n    \"userKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\",\n    \"userKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the RevokePermissionsFromUser1 tool (Status: 400, Content-Type: application/json)
const RevokePermissionsFromUser1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if incorrect permissions are passed in request (for e.g. non existing operation or global permission).\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the RevokePermissionsFromUser1 tool (Status: 401, Content-Type: application/json)
const RevokePermissionsFromUser1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the RevokePermissionsFromUser1 tool (Status: 403, Content-Type: application/json)
const RevokePermissionsFromUser1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the RevokePermissionsFromUser1 tool (Status: 404, Content-Type: application/json)
const RevokePermissionsFromUser1ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if user with specified key not found.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewRevokePermissionsFromUser1MCPTool creates the MCP Tool instance for RevokePermissionsFromUser1
func NewRevokePermissionsFromUser1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RevokePermissionsFromUser1",
		"Revoke space permissions from a user - Revoke permissions from a user in the given space.\nIf user doesn't have permissions that we are trying to revoke, those permissions will be silently skipped.\nMultiple permissions could be passed in one request. Supported targetType and operationKey pairs:\n* space read\n* space administer\n* space export\n* space restrict\n* space delete_own\n* space delete_mail\n* page create\n* page delete\n* blogpost create\n* blogpost delete\n* comment create\n* comment delete\n* attachment create\n* attachment delete\n\nSee <a href=\"https://confluence.atlassian.com/display/DOC/Space+Permissions+Overview\">Space Permissions documentation</a> for additional information about supported permissions.\n\nExample request URI's:\n"+"\x60"+"https://example.com/confluence/rest/api/space/TESTSPACE/permissions/user/4028ae289154667d0191546bd5840000/revoke"+"\x60"+"",
		[]byte(RevokePermissionsFromUser1InputSchema),
	)
}

// RevokePermissionsFromUser1Handler is the handler function for the RevokePermissionsFromUser1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RevokePermissionsFromUser1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/space/{spaceKey}/permissions/user/{userKey}/revoke", args, []string{"spaceKey", "userKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RevokePermissionsFromUser1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
