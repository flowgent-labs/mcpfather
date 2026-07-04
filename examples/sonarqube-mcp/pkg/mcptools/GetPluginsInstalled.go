package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetPluginsInstalled tool
const GetPluginsInstalledInputSchema = "{\n  \"properties\": {\n    \"f\": {\n      \"description\": \"Comma-separated list of the additional fields to be returned in response. No additional field is returned by default. Possible values are: - category - category as defined in the Update Center. A connection to the Update Center is needed\",\n      \"enum\": [\n        \"category\"\n      ],\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"description\": \"Allows to filter plugins by type\",\n      \"enum\": [\n        \"BUNDLED\",\n        \"EXTERNAL\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPluginsInstalled tool (Status: 200, Content-Type: application/json)
const GetPluginsInstalledResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetPluginsInstalledMCPTool creates the MCP Tool instance for GetPluginsInstalled
func NewGetPluginsInstalledMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPluginsInstalled",
		"Get the list of all the plugins installed on the SonarQube instance, sorted by plugin name. - Get the list of all the plugins installed on the SonarQube instance, sorted by plugin name. Requires authentication.",
		[]byte(GetPluginsInstalledInputSchema),
	)
}

// GetPluginsInstalledHandler is the handler function for the GetPluginsInstalled tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPluginsInstalledHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/plugins/installed", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPluginsInstalled")
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
