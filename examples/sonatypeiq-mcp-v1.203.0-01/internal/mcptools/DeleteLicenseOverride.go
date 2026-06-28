package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteLicenseOverride tool
const DeleteLicenseOverrideInputSchema = "{\n  \"properties\": {\n    \"licenseOverrideId\": {\n      \"description\": \"Enter the id of the license override you want to delete.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the id of the application, organization or the repository.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the " + "\x60" + "ownerType" + "\x60" + " scope for which you want to delete license override\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    },\n    \"where\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"licenseOverrideId\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteLicenseOverrideMCPTool creates the MCP Tool instance for DeleteLicenseOverride
func NewDeleteLicenseOverrideMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteLicenseOverride",
		"Use this method to delete a license override for a component.\n\nPermissions required: Change Licenses",
		[]byte(DeleteLicenseOverrideInputSchema),
	)
}

// DeleteLicenseOverrideHandler is the handler function for the DeleteLicenseOverride tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteLicenseOverrideHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/licenseOverrides/{ownerType}/{ownerId}/{licenseOverrideId}", args, []string{"licenseOverrideId", "ownerId", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteLicenseOverride")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
