package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the ArchiveProject tool
const ArchiveProjectInputSchema = "{\n  \"properties\": {\n    \"projectIdOrKey\": {\n      \"description\": \"Project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewArchiveProjectMCPTool creates the MCP Tool instance for ArchiveProject
func NewArchiveProjectMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ArchiveProject",
		"Archive a project - Archives a project",
		[]byte(ArchiveProjectInputSchema),
	)
}

// ArchiveProjectHandler is the handler function for the ArchiveProject tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ArchiveProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/project/{projectIdOrKey}/archive", args, []string{"projectIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ArchiveProject")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
