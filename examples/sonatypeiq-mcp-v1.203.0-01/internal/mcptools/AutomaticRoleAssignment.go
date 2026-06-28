package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AutomaticRoleAssignment tool
const AutomaticRoleAssignmentInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"mappings\": {\n          \"items\": {\n            \"properties\": {\n              \"from\": {\n                \"enum\": [\n                  \"SCM_USERNAME\",\n                  \"SCM_EMAIL\",\n                  \"SCM_FULLNAME\",\n                  \"GITLOG_EMAIL\",\n                  \"GITLOG_FULLNAME\"\n                ],\n                \"type\": \"string\"\n              },\n              \"to\": {\n                \"enum\": [\n                  \"IQ_USERNAME\",\n                  \"IQ_EMAIL\",\n                  \"IQ_FULLNAME\"\n                ],\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        },\n        \"role\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"publicId\": {\n      \"description\": \"Enter the public applicationId for automatic role assignment.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"publicId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AutomaticRoleAssignment tool (Status: 200, Content-Type: application/json)
const AutomaticRoleAssignmentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The 'developer' role has automatically been assigned to all contributors of the repository, who matched IQ Server users via the provided matching strategies.\n\nThe response contains all usernames that were successfully granted the role provided on the given application as well as an indication of which matching strategy was the first to match a user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **matchedUsers** (Type: array):\n      - Unique Items: true\n    - **Items** (Type: string):\n  - **successfulMapping** (Type: object):\n    - **to** (Type: string):\n        - Enum: ['IQ_USERNAME', 'IQ_EMAIL', 'IQ_FULLNAME']\n    - **from** (Type: string):\n        - Enum: ['SCM_USERNAME', 'SCM_EMAIL', 'SCM_FULLNAME', 'GITLOG_EMAIL', 'GITLOG_FULLNAME']\n"

// NewAutomaticRoleAssignmentMCPTool creates the MCP Tool instance for AutomaticRoleAssignment
func NewAutomaticRoleAssignmentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AutomaticRoleAssignment",
		"Use this method to automatically grant the supplied role to all contributors of a repository on a given application.\n\nPrerequisites for automatic role assignment are:<ol><li>SCM configuration for the application and authentication token should exist.</li><li>The contributors to the repository should match a user in IQ based on the supplied mappings.</li><li>Either user mapping strategies have been configured for your organization, or they are provided in the request</li></ol>\n\nPermissions required: Edit access control on the application.",
		[]byte(AutomaticRoleAssignmentInputSchema),
	)
}

// AutomaticRoleAssignmentHandler is the handler function for the AutomaticRoleAssignment tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AutomaticRoleAssignmentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "*/*"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/sourceControl/automaticRoleAssignment/{publicId}", args, []string{"publicId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AutomaticRoleAssignment")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
