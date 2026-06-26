package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the SetPermissions1 tool
const SetPermissions1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"items\": {\n        \"properties\": {\n          \"groupName\": {\n            \"type\": \"string\"\n          },\n          \"operations\": {\n            \"items\": {\n              \"properties\": {\n                \"operationKey\": {\n                  \"example\": \"read\",\n                  \"type\": \"string\"\n                },\n                \"targetType\": {\n                  \"example\": \"space\",\n                  \"type\": \"string\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"type\": \"array\",\n            \"uniqueItems\": true\n          },\n          \"userKey\": {\n            \"type\": \"string\"\n          }\n        },\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    },\n    \"spaceKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the SetPermissions1 tool (Status: 400, Content-Type: application/json)
const SetPermissions1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if incorrect permissions are passed in request (for e.g. non existing operation or global permission).\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the SetPermissions1 tool (Status: 401, Content-Type: application/json)
const SetPermissions1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the SetPermissions1 tool (Status: 403, Content-Type: application/json)
const SetPermissions1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the SetPermissions1 tool (Status: 404, Content-Type: application/json)
const SetPermissions1ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if user or group with specified key not found.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewSetPermissions1MCPTool creates the MCP Tool instance for SetPermissions1
func NewSetPermissions1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetPermissions1",
		"Set permissions to multiple users/groups/anonymous user in the given space - Sets permissions to multiple users/groups in the given space.\nRequest should contain all permissions that user/group/anonymous user will have in a given space.\nIf permission is absent in the request, but was granted before, it will be revoked.\nIf empty list of permissions passed to user/group/anonymous user, then all their existing permissions will be revoked.\nIf user/group/anonymous user not mentioned in the request, their permissions will not be revoked.\n\nMaximum 40 different users/groups/anonymous user could be passed in the request.\n\nSee <a href=\"https://confluence.atlassian.com/display/DOC/Space+Permissions+Overview\">Space Permissions documentation</a> for additional information about supported permissions.\n\nExample request URI's:\n"+"\x60"+"https://example.com/confluence/rest/api/space/TESTSPACE/permissions"+"\x60"+"",
		[]byte(SetPermissions1InputSchema),
	)
}

// SetPermissions1Handler is the handler function for the SetPermissions1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetPermissions1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/space/{spaceKey}/permissions", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetPermissions1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
