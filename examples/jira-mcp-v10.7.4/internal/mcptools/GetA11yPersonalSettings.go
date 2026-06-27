package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetA11yPersonalSettings tool
const GetA11yPersonalSettingsInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetA11yPersonalSettings tool (Status: 200, Content-Type: application/json)
const GetA11yPersonalSettingsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned when validation succeeded.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **key** (Type: string):\n      - Example: 'a11y-setting-underlined-links'\n  - **enabled** (Type: boolean):\n"

// NewGetA11yPersonalSettingsMCPTool creates the MCP Tool instance for GetA11yPersonalSettings
func NewGetA11yPersonalSettingsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetA11yPersonalSettings",
		"Get available accessibility personal settings - Returns available accessibility personal settings along with "+"\x60"+"enabled"+"\x60"+" property that indicates the currently logged-in user preference.",
		[]byte(GetA11yPersonalSettingsInputSchema),
	)
}

// GetA11yPersonalSettingsHandler is the handler function for the GetA11yPersonalSettings tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetA11yPersonalSettingsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/a11y/personal-settings", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetA11yPersonalSettings"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
