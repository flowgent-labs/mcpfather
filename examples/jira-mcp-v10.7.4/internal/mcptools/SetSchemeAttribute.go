package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetSchemeAttribute tool
const SetSchemeAttributeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"The key of the permission scheme attribute.\",\n      \"type\": \"string\"\n    },\n    \"permissionSchemeId\": {\n      \"description\": \"The id of the permission scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"permissionSchemeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetSchemeAttributeMCPTool creates the MCP Tool instance for SetSchemeAttribute
func NewSetSchemeAttributeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetSchemeAttribute",
		"Update or insert a scheme attribute - Updates or inserts the attribute for a permission scheme specified by permission scheme id. The attribute consists of the key and the value. The value will be converted to Boolean using Boolean#valueOf.",
		[]byte(SetSchemeAttributeInputSchema),
	)
}

// SetSchemeAttributeHandler is the handler function for the SetSchemeAttribute tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetSchemeAttributeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "text/plain"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/permissionscheme/{permissionSchemeId}/attribute/{key}", args, []string{"key", "permissionSchemeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetSchemeAttribute"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
