package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostQualityprofilesRename tool
const PostQualityprofilesRenameInputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"Quality profile key\",\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"New quality profile name\",\n      \"maxLength\": 100,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"name\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostQualityprofilesRenameMCPTool creates the MCP Tool instance for PostQualityprofilesRename
func NewPostQualityprofilesRenameMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostQualityprofilesRename",
		"Rename a quality profile. - Rename a quality profile. Requires one of the following permissions: - 'Administer Quality Profiles' - Edit right on the specified quality profile",
		[]byte(PostQualityprofilesRenameInputSchema),
	)
}

// PostQualityprofilesRenameHandler is the handler function for the PostQualityprofilesRename tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostQualityprofilesRenameHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/qualityprofiles/rename", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostQualityprofilesRename")
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
