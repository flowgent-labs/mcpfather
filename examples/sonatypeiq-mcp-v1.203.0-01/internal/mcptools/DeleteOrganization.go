package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteOrganization tool
const DeleteOrganizationInputSchema = "{\n  \"properties\": {\n    \"organizationId\": {\n      \"description\": \"Enter the organization id to be deleted.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteOrganizationMCPTool creates the MCP Tool instance for DeleteOrganization
func NewDeleteOrganizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteOrganization",
		"Use this method to delete an existing organization, by providing the organization id.\n\nPermissions required: Edit IQ Elements",
		[]byte(DeleteOrganizationInputSchema),
	)
}

// DeleteOrganizationHandler is the handler function for the DeleteOrganization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteOrganizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/organizations/{organizationId}", args, []string{"organizationId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteOrganization"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
