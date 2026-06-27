package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the RemovePreference tool
const RemovePreferenceInputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"Key of the preference to be removed.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewRemovePreferenceMCPTool creates the MCP Tool instance for RemovePreference
func NewRemovePreferenceMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RemovePreference",
		"Delete user preference - Removes preference of the currently logged in user. Preference key must be provided as input parameters (key). If key parameter is not provided or wrong - status code 404. If preference is unset - status code 204.",
		[]byte(RemovePreferenceInputSchema),
	)
}

// RemovePreferenceHandler is the handler function for the RemovePreference tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RemovePreferenceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/mypreferences", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RemovePreference"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
