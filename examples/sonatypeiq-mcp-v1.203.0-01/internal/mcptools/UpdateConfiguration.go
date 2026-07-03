package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the UpdateConfiguration tool
const UpdateConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"userTokenDefaultExpirationDays\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the UpdateConfiguration tool (Status: 200, Content-Type: application/json)
const UpdateConfigurationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Configuration updated successfully\n\n## Response Structure\n\n- Structure (Type: object):\n  - **userTokenDefaultExpirationDays** (Type: integer, int32):\n"

// NewUpdateConfigurationMCPTool creates the MCP Tool instance for UpdateConfiguration
func NewUpdateConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateConfiguration",
		"Use this method to update user token configuration. Null values are ignored (no change). Returns the current configuration after applying changes.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(UpdateConfigurationInputSchema),
	)
}

// UpdateConfigurationHandler is the handler function for the UpdateConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/userTokens", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateConfiguration")
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
