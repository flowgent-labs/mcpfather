package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DisableFeature tool
const DisableFeatureInputSchema = "{\n  \"properties\": {\n    \"feature\": {\n      \"description\": \"Enter the name of the IQ Server feature to be disabled.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"feature\"\n  ],\n  \"type\": \"object\"\n}"

// NewDisableFeatureMCPTool creates the MCP Tool instance for DisableFeature
func NewDisableFeatureMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DisableFeature",
		"Use this method to disable an IQ Server feature.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(DisableFeatureInputSchema),
	)
}

// DisableFeatureHandler is the handler function for the DisableFeature tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DisableFeatureHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/config/features/{feature}", args, []string{"feature"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DisableFeature")
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
