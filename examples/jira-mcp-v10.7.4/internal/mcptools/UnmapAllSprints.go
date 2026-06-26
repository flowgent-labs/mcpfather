package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UnmapAllSprints tool
const UnmapAllSprintsInputSchema = "{\n  \"type\": \"object\"\n}"

// NewUnmapAllSprintsMCPTool creates the MCP Tool instance for UnmapAllSprints
func NewUnmapAllSprintsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UnmapAllSprints",
		"Unmap all sprints from being synced - Sets the Synced flag to false for all sprints on this Jira instance. This operation is intended for cleanup only. It is highly destructive and not reversible. Use with caution.",
		[]byte(UnmapAllSprintsInputSchema),
	)
}

// UnmapAllSprintsHandler is the handler function for the UnmapAllSprints tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UnmapAllSprintsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/agile/1.0/sprint/unmap-all", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UnmapAllSprints"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
