package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the Delete1 tool
const Delete1InputSchema = "{\n  \"properties\": {\n    \"realm\": {\n      \"default\": \"Internal\",\n      \"description\": \"Enter the " + "\x60" + "realm" + "\x60" + ". Allowed values are " + "\x60" + "Internal" + "\x60" + "," + "\x60" + "OAUTH2" + "\x60" + ", and " + "\x60" + "SAML" + "\x60" + ".\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"Enter the username to be deleted.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// NewDelete1MCPTool creates the MCP Tool instance for Delete1
func NewDelete1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Delete1",
		"Use this method to delete an existing user.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(Delete1InputSchema),
	)
}

// Delete1Handler is the handler function for the Delete1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Delete1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/users/{username}", args, []string{"username"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Delete1")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
