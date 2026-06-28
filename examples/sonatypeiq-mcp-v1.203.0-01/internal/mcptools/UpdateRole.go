package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the UpdateRole tool
const UpdateRoleInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"builtIn\": {\n          \"type\": \"boolean\"\n        },\n        \"description\": {\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"permissionCategories\": {\n          \"items\": {\n            \"properties\": {\n              \"displayName\": {\n                \"type\": \"string\"\n              },\n              \"permissions\": {\n                \"items\": {\n                  \"properties\": {\n                    \"allowed\": {\n                      \"type\": \"boolean\"\n                    },\n                    \"description\": {\n                      \"type\": \"string\"\n                    },\n                    \"displayName\": {\n                      \"type\": \"string\"\n                    },\n                    \"id\": {\n                      \"enum\": [\n                        \"CONFIGURE_SYSTEM\",\n                        \"EDIT_ROLES\",\n                        \"VIEW_ROLES\",\n                        \"ACCESS_AUDIT_LOG\",\n                        \"WAIVE_POLICY_VIOLATIONS\",\n                        \"CHANGE_LICENSES\",\n                        \"CHANGE_SECURITY_VULNERABILITIES\",\n                        \"MANAGE_PROPRIETARY\",\n                        \"CLAIM_COMPONENT\",\n                        \"WRITE\",\n                        \"READ\",\n                        \"EDIT_ACCESS_CONTROL\",\n                        \"EVALUATE_APPLICATION\",\n                        \"EVALUATE_COMPONENT\",\n                        \"ADD_APPLICATION\",\n                        \"MANAGE_AUTOMATIC_APPLICATION_CREATION\",\n                        \"MANAGE_AUTOMATIC_SCM_CONFIGURATION\",\n                        \"LEGAL_REVIEWER\",\n                        \"CREATE_PULL_REQUESTS\"\n                      ],\n                      \"type\": \"string\"\n                    }\n                  },\n                  \"type\": \"object\"\n                },\n                \"type\": \"array\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"roleId\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"roleId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateRole tool (Status: 200, Content-Type: application/json)
const UpdateRoleResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The role was updated successfully. The response contains the updated role details.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **builtIn** (Type: boolean):\n  - **description** (Type: string):\n  - **id** (Type: string):\n  - **name** (Type: string):\n  - **permissionCategories** (Type: array):\n    - **Items** (Type: object):\n      - **displayName** (Type: string):\n      - **permissions** (Type: array):\n        - **Items** (Type: object):\n          - **description** (Type: string):\n          - **displayName** (Type: string):\n          - **id** (Type: string):\n              - Enum: ['CONFIGURE_SYSTEM', 'EDIT_ROLES', 'VIEW_ROLES', 'ACCESS_AUDIT_LOG', 'WAIVE_POLICY_VIOLATIONS', 'CHANGE_LICENSES', 'CHANGE_SECURITY_VULNERABILITIES', 'MANAGE_PROPRIETARY', 'CLAIM_COMPONENT', 'WRITE', 'READ', 'EDIT_ACCESS_CONTROL', 'EVALUATE_APPLICATION', 'EVALUATE_COMPONENT', 'ADD_APPLICATION', 'MANAGE_AUTOMATIC_APPLICATION_CREATION', 'MANAGE_AUTOMATIC_SCM_CONFIGURATION', 'LEGAL_REVIEWER', 'CREATE_PULL_REQUESTS']\n          - **allowed** (Type: boolean):\n"

// NewUpdateRoleMCPTool creates the MCP Tool instance for UpdateRole
func NewUpdateRoleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateRole",
		"Use this method to update an existing custom role and its permissions.\n\nPermissions required: Edit Roles",
		[]byte(UpdateRoleInputSchema),
	)
}

// UpdateRoleHandler is the handler function for the UpdateRole tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/roles/{roleId}", args, []string{"roleId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateRole")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
