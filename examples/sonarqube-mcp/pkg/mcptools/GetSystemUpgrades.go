package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetSystemUpgrades tool
const GetSystemUpgradesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetSystemUpgrades tool (Status: 200, Content-Type: application/json)
const GetSystemUpgradesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetSystemUpgradesMCPTool creates the MCP Tool instance for GetSystemUpgrades
func NewGetSystemUpgradesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSystemUpgrades",
		"Lists available upgrades for the SonarQube instance (if any) and for each one, lists incompatible plugins and plugins requiring upgrade. - Lists available upgrades for the SonarQube instance (if any) and for each one, lists incompatible plugins and plugins requiring upgrade. Plugin information is retrieved from Update Center. Date and time at which Update Center was last refreshed is provided in the response.",
		[]byte(GetSystemUpgradesInputSchema),
	)
}

// GetSystemUpgradesHandler is the handler function for the GetSystemUpgrades tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSystemUpgradesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/system/upgrades", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSystemUpgrades")
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
