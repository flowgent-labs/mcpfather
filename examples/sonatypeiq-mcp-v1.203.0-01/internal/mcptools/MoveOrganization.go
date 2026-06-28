package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the MoveOrganization tool
const MoveOrganizationInputSchema = "{\n  \"properties\": {\n    \"destinationId\": {\n      \"description\": \"Enter the id for the new parent organization.\",\n      \"type\": \"string\"\n    },\n    \"failEarlyOnError\": {\n      \"default\": false,\n      \"type\": \"boolean\"\n    },\n    \"organizationId\": {\n      \"description\": \"Enter the id for the organization to be moved under the new parent.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"destinationId\",\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the MoveOrganization tool (Status: 200, Content-Type: application/json)
const MoveOrganizationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The organization has been successfully moved under the parent organization id provided.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **errors** (Type: array):\n    - **Items** (Type: object):\n      - **message** (Type: string):\n      - **type** (Type: string):\n          - Enum: ['TAG', 'POLICY', 'LICENSE_THREAT_GROUP', 'LABEL', 'PARENT_HIERARCHY']\n  - **warnings** (Type: array):\n    - **Items** (Type: object):\n      - **message** (Type: string):\n      - **type** (Type: string):\n          - Enum: ['LICENSE_OVERRIDE', 'POLICY_MONITORING', 'POLICY_WAIVER']\n"

// Response Template for the MoveOrganization tool (Status: 409, Content-Type: application/json)
const MoveOrganizationResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Encountered conflicts while inheriting policy elements of the new parent organization. The organization could not be moved under the new parent organization id provided.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **errors** (Type: array):\n    - **Items** (Type: object):\n      - **message** (Type: string):\n      - **type** (Type: string):\n          - Enum: ['TAG', 'POLICY', 'LICENSE_THREAT_GROUP', 'LABEL', 'PARENT_HIERARCHY']\n  - **warnings** (Type: array):\n    - **Items** (Type: object):\n      - **message** (Type: string):\n      - **type** (Type: string):\n          - Enum: ['LICENSE_OVERRIDE', 'POLICY_MONITORING', 'POLICY_WAIVER']\n"

// NewMoveOrganizationMCPTool creates the MCP Tool instance for MoveOrganization
func NewMoveOrganizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"MoveOrganization",
		"Use this method to change the parent organization.\n\nPermissions required: Edit IQ Elements",
		[]byte(MoveOrganizationInputSchema),
	)
}

// MoveOrganizationHandler is the handler function for the MoveOrganization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveOrganizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/organizations/{organizationId}/move/destination/{destinationId}", args, []string{"destinationId", "organizationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "MoveOrganization")
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
