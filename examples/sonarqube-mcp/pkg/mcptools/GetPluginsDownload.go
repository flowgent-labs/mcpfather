package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetPluginsDownload tool
const GetPluginsDownloadInputSchema = "{\n  \"properties\": {\n    \"plugin\": {\n      \"description\": \"The key identifying the plugin to download\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"plugin\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPluginsDownload tool (Status: 200, Content-Type: application/json)
const GetPluginsDownloadResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewGetPluginsDownloadMCPTool creates the MCP Tool instance for GetPluginsDownload
func NewGetPluginsDownloadMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPluginsDownload",
		"Download plugin JAR, for usage by scanner engine. - Download plugin JAR, for usage by scanner engine",
		[]byte(GetPluginsDownloadInputSchema),
	)
}

// GetPluginsDownloadHandler is the handler function for the GetPluginsDownload tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPluginsDownloadHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/plugins/download", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPluginsDownload")
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
