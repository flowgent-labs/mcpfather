package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostQualitygatesSetAsDefault tool
const PostQualitygatesSetAsDefaultInputSchema = "{\n  \"properties\": {\n    \"name\": {\n      \"description\": \"Name of the quality gate to set as default\",\n      \"maxLength\": 100,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"name\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostQualitygatesSetAsDefaultMCPTool creates the MCP Tool instance for PostQualitygatesSetAsDefault
func NewPostQualitygatesSetAsDefaultMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostQualitygatesSetAsDefault",
		"Set a quality gate as the default quality gate. - Set a quality gate as the default quality gate. Parameter 'name' must be specified. Requires the 'Administer Quality Gates' permission.",
		[]byte(PostQualitygatesSetAsDefaultInputSchema),
	)
}

// PostQualitygatesSetAsDefaultHandler is the handler function for the PostQualitygatesSetAsDefault tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostQualitygatesSetAsDefaultHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/qualitygates/set_as_default", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostQualitygatesSetAsDefault")
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
