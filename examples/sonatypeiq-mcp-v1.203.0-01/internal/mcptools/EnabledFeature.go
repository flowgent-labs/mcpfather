package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the EnabledFeature tool
const EnabledFeatureInputSchema = "{\n  \"properties\": {\n    \"feature\": {\n      \"description\": \"Enter the name of the feature to be enabled.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"feature\"\n  ],\n  \"type\": \"object\"\n}"

// NewEnabledFeatureMCPTool creates the MCP Tool instance for EnabledFeature
func NewEnabledFeatureMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"EnabledFeature",
		"Use this method to enable an IQ Server feature.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(EnabledFeatureInputSchema),
	)
}

// EnabledFeatureHandler is the handler function for the EnabledFeature tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func EnabledFeatureHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/config/features/{feature}", args, []string{"feature"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "EnabledFeature")
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
