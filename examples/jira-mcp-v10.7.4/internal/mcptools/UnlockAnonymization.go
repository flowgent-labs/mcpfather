package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UnlockAnonymization tool
const UnlockAnonymizationInputSchema = "{\n  \"type\": \"object\"\n}"

// NewUnlockAnonymizationMCPTool creates the MCP Tool instance for UnlockAnonymization
func NewUnlockAnonymizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UnlockAnonymization",
		"Delete stale user anonymization task - Removes stale user anonymization task, for scenarios when the node that was executing it is no longer alive. Use it only after making sure that the parent node of the task is actually down, and not just having connectivity issues.",
		[]byte(UnlockAnonymizationInputSchema),
	)
}

// UnlockAnonymizationHandler is the handler function for the UnlockAnonymization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UnlockAnonymizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/user/anonymization/unlock", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UnlockAnonymization"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
