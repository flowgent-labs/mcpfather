package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetTemplateForNewRole tool
const GetTemplateForNewRoleInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetTemplateForNewRole tool (Status: 200, Content-Type: application/json)
const GetTemplateForNewRoleResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a role template with available permissions.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n  - **id** (Type: string):\n  - **name** (Type: string):\n  - **permissionCategories** (Type: array):\n    - **Items** (Type: object):\n      - **displayName** (Type: string):\n      - **permissions** (Type: array):\n        - **Items** (Type: object):\n          - **displayName** (Type: string):\n          - **id** (Type: string):\n              - Enum: ['CONFIGURE_SYSTEM', 'EDIT_ROLES', 'VIEW_ROLES', 'ACCESS_AUDIT_LOG', 'WAIVE_POLICY_VIOLATIONS', 'CHANGE_LICENSES', 'CHANGE_SECURITY_VULNERABILITIES', 'MANAGE_PROPRIETARY', 'CLAIM_COMPONENT', 'WRITE', 'READ', 'EDIT_ACCESS_CONTROL', 'EVALUATE_APPLICATION', 'EVALUATE_COMPONENT', 'ADD_APPLICATION', 'MANAGE_AUTOMATIC_APPLICATION_CREATION', 'MANAGE_AUTOMATIC_SCM_CONFIGURATION', 'LEGAL_REVIEWER', 'CREATE_PULL_REQUESTS']\n          - **allowed** (Type: boolean):\n          - **description** (Type: string):\n  - **builtIn** (Type: boolean):\n"

// NewGetTemplateForNewRoleMCPTool creates the MCP Tool instance for GetTemplateForNewRole
func NewGetTemplateForNewRoleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetTemplateForNewRole",
		"Use this method to retrieve a template for creating a new custom role, including all available permissions that can be assigned.\n\nPermissions required: Edit Roles",
		[]byte(GetTemplateForNewRoleInputSchema),
	)
}

// GetTemplateForNewRoleHandler is the handler function for the GetTemplateForNewRole tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetTemplateForNewRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/roles/new", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetTemplateForNewRole")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
