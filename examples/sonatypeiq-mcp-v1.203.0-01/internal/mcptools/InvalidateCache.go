package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the InvalidateCache tool
const InvalidateCacheInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the InvalidateCache tool (Status: 200, Content-Type: application/json)
const InvalidateCacheResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Cache cleared successfully. Returns the number of entries that were invalidated.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **entriesInvalidated** (Type: integer, int64):\n"

// NewInvalidateCacheMCPTool creates the MCP Tool instance for InvalidateCache
func NewInvalidateCacheMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"InvalidateCache",
		"Clear the integration version cache. Use this endpoint after a new integration version is released to ensure IQ Server immediately recognizes the new version instead of waiting for cache expiration (10 minutes).\n\nPermissions required: Edit System Configuration and Users",
		[]byte(InvalidateCacheInputSchema),
	)
}

// InvalidateCacheHandler is the handler function for the InvalidateCache tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func InvalidateCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/config/integrationVersions/cache", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "InvalidateCache")
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
