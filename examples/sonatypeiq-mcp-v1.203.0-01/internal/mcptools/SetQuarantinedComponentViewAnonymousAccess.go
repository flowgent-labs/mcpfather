package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetQuarantinedComponentViewAnonymousAccess tool
const SetQuarantinedComponentViewAnonymousAccessInputSchema = "{\n  \"properties\": {\n    \"enabled\": {\n      \"description\": \"Select " + "\x60" + "true" + "\x60" + " or " + "\x60" + "false" + "\x60" + " to enable or disable anonymous access.\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"required\": [\n    \"enabled\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetQuarantinedComponentViewAnonymousAccessMCPTool creates the MCP Tool instance for SetQuarantinedComponentViewAnonymousAccess
func NewSetQuarantinedComponentViewAnonymousAccessMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetQuarantinedComponentViewAnonymousAccess",
		"Use this method to enable/disable anonymous access to view the quarantined components.\n\nPermissions required: Edit IQ Elements",
		[]byte(SetQuarantinedComponentViewAnonymousAccessInputSchema),
	)
}

// SetQuarantinedComponentViewAnonymousAccessHandler is the handler function for the SetQuarantinedComponentViewAnonymousAccess tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetQuarantinedComponentViewAnonymousAccessHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/firewall/quarantinedComponentView/configuration/anonymousAccess/{enabled}", args, []string{"enabled"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetQuarantinedComponentViewAnonymousAccess")
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
