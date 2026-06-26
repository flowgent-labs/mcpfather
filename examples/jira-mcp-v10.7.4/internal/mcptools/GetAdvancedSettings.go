package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAdvancedSettings tool
const GetAdvancedSettingsInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAdvancedSettings tool (Status: 200, Content-Type: application/json)
const GetAdvancedSettingsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns all properties to display in the \"General Configuration > Advanced Settings\" page.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **example** (Type: string):\n  - **key** (Type: string):\n  - **value** (Type: string):\n"

// NewGetAdvancedSettingsMCPTool creates the MCP Tool instance for GetAdvancedSettings
func NewGetAdvancedSettingsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAdvancedSettings",
		"Get all advanced settings properties - Returns the properties that are displayed on the \"General Configuration > Advanced Settings\" page.",
		[]byte(GetAdvancedSettingsInputSchema),
	)
}

// GetAdvancedSettingsHandler is the handler function for the GetAdvancedSettings tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAdvancedSettingsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/application-properties/advanced-settings", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAdvancedSettings"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
