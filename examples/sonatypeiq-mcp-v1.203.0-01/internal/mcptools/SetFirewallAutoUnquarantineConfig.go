package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetFirewallAutoUnquarantineConfig tool
const SetFirewallAutoUnquarantineConfigInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Enter value for each repository and the required status for auto-release as " + "\x60" + "true" + "\x60" + " or " + "\x60" + "false" + "\x60" + ".\",\n      \"items\": {\n        \"properties\": {\n          \"autoReleaseQuarantineEnabled\": {\n            \"type\": \"boolean\"\n          },\n          \"id\": {\n            \"type\": \"string\"\n          },\n          \"name\": {\n            \"type\": \"string\"\n          }\n        },\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the SetFirewallAutoUnquarantineConfig tool (Status: 200, Content-Type: application/json)
const SetFirewallAutoUnquarantineConfigResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains each updated " + "\x60" + "autoReleaseQuarantineEnabled" + "\x60" + " status for the repositories requested.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **autoReleaseQuarantineEnabled** (Type: boolean):\n    - **id** (Type: string):\n    - **name** (Type: string):\n"

// NewSetFirewallAutoUnquarantineConfigMCPTool creates the MCP Tool instance for SetFirewallAutoUnquarantineConfig
func NewSetFirewallAutoUnquarantineConfigMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetFirewallAutoUnquarantineConfig",
		"Use this method to set the configurations for auto-release from quarantine for a list of repositories.\n\nPermissions required: Edit IQ Elements",
		[]byte(SetFirewallAutoUnquarantineConfigInputSchema),
	)
}

// SetFirewallAutoUnquarantineConfigHandler is the handler function for the SetFirewallAutoUnquarantineConfig tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetFirewallAutoUnquarantineConfigHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/firewall/releaseQuarantine/configuration", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetFirewallAutoUnquarantineConfig")
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
