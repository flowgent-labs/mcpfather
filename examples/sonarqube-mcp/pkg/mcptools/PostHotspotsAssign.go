package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostHotspotsAssign tool
const PostHotspotsAssignInputSchema = "{\n  \"properties\": {\n    \"assignee\": {\n      \"description\": \"Login of the assignee with 'Browse' project permission\",\n      \"type\": \"string\"\n    },\n    \"comment\": {\n      \"description\": \"A comment provided with assign action\",\n      \"type\": \"string\"\n    },\n    \"hotspot\": {\n      \"description\": \"Hotspot key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"hotspot\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostHotspotsAssignMCPTool creates the MCP Tool instance for PostHotspotsAssign
func NewPostHotspotsAssignMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostHotspotsAssign",
		"Assign a hotspot to an active user. - Assign a hotspot to an active user. Requires authentication and Browse permission on project",
		[]byte(PostHotspotsAssignInputSchema),
	)
}

// PostHotspotsAssignHandler is the handler function for the PostHotspotsAssign tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostHotspotsAssignHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/hotspots/assign", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostHotspotsAssign")
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
