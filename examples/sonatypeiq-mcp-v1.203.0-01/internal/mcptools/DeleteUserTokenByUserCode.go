package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteUserTokenByUserCode tool
const DeleteUserTokenByUserCodeInputSchema = "{\n  \"properties\": {\n    \"userCode\": {\n      \"description\": \"Enter the " + "\x60" + "userCode" + "\x60" + " to be deleted.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"userCode\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteUserTokenByUserCodeMCPTool creates the MCP Tool instance for DeleteUserTokenByUserCode
func NewDeleteUserTokenByUserCodeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteUserTokenByUserCode",
		"Use this method to delete an existing user token by specifying the userCode.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(DeleteUserTokenByUserCodeInputSchema),
	)
}

// DeleteUserTokenByUserCodeHandler is the handler function for the DeleteUserTokenByUserCode tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteUserTokenByUserCodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/userTokens/userCode/{userCode}", args, []string{"userCode"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteUserTokenByUserCode")
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
