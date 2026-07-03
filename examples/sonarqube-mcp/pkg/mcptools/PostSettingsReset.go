package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostSettingsReset tool
const PostSettingsResetInputSchema = "{\n  \"properties\": {\n    \"component\": {\n      \"description\": \"Component key. Only keys for projects, applications, portfolios or subportfolios are accepted.\",\n      \"type\": \"string\"\n    },\n    \"keys\": {\n      \"description\": \"Comma-separated list of keys\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"keys\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostSettingsResetMCPTool creates the MCP Tool instance for PostSettingsReset
func NewPostSettingsResetMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostSettingsReset",
		"Remove a setting value. - Remove a setting value. The settings defined in conf/sonar.properties are read-only and can't be changed. Requires one of the following permissions: - 'Administer System' - 'Administer' rights on the specified component",
		[]byte(PostSettingsResetInputSchema),
	)
}

// PostSettingsResetHandler is the handler function for the PostSettingsReset tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostSettingsResetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/settings/reset", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostSettingsReset")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
