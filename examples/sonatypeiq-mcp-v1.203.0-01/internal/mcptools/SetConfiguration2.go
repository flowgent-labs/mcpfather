package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetConfiguration2 tool
const SetConfiguration2InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Enter the Jira configuration details here.\",\n      \"properties\": {\n        \"customFields\": {\n          \"additionalProperties\": {\n            \"type\": \"object\"\n          },\n          \"type\": \"object\"\n        },\n        \"password\": {\n          \"type\": \"string\"\n        },\n        \"url\": {\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewSetConfiguration2MCPTool creates the MCP Tool instance for SetConfiguration2
func NewSetConfiguration2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetConfiguration2",
		"Use this method to set a Jira configuration. If a Jira configuration already exists, the values will be updated with the ones provided here. If the server URL is being changed, then the password (if any) will be required.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(SetConfiguration2InputSchema),
	)
}

// SetConfiguration2Handler is the handler function for the SetConfiguration2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetConfiguration2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/jira", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetConfiguration2")
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
