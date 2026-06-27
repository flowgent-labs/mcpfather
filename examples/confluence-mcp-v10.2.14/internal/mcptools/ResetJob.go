package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the ResetJob tool
const ResetJobInputSchema = "{\n  \"type\": \"object\"\n}"

// NewResetJobMCPTool creates the MCP Tool instance for ResetJob
func NewResetJobMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ResetJob",
		"Reset reindex job status - Resets the status of the current in-progress reindex job.\nThis is useful when a reindex job has failed/stalled and needs to be cleared before starting a new reindex operation.\nThis operation is only available to system administrators.\n",
		[]byte(ResetJobInputSchema),
	)
}

// ResetJobHandler is the handler function for the ResetJob tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ResetJobHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/index/resetjob", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ResetJob"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
