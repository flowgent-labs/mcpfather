package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetSystemHealth tool
const GetSystemHealthInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetSystemHealth tool (Status: 200, Content-Type: application/json)
const GetSystemHealthResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetSystemHealthMCPTool creates the MCP Tool instance for GetSystemHealth
func NewGetSystemHealthMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSystemHealth",
		"Provide health status of SonarQube. - Provide health status of SonarQube. Although global health is calculated based on both application and search nodes, detailed information is returned only for application nodes. - GREEN: SonarQube is fully operational - YELLOW: SonarQube is usable, but it needs attention in order to be fully operational - RED: SonarQube is not operational Requires the 'Administer System' permission or system passcode (see sonar.web.systemPasscode in sonar.properties). When SonarQube is in safe mode (waiting or running a database upgrade), only the authentication with a system passcode is supported.",
		[]byte(GetSystemHealthInputSchema),
	)
}

// GetSystemHealthHandler is the handler function for the GetSystemHealth tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSystemHealthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/system/health", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSystemHealth")
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
