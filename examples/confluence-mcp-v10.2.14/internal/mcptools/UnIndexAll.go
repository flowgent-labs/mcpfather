package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UnIndexAll tool
const UnIndexAllInputSchema = "{\n  \"type\": \"object\"\n}"

// NewUnIndexAllMCPTool creates the MCP Tool instance for UnIndexAll
func NewUnIndexAllMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UnIndexAll",
		"Remove all content from search index - Removes all content from the search index, effectively clearing the entire search index.\nThis operation is destructive and will require a full reindex to restore search functionality.\nThis operation is only available to system administrators.\n\n**Warning**: This operation will remove all searchable content from the index.\nUsers will not be able to search for content until a reindex is performed.\n\nExample request URI:\n"+"\x60"+"http://example.com/confluence/rest/api/index/unindex"+"\x60"+"\n",
		[]byte(UnIndexAllInputSchema),
	)
}

// UnIndexAllHandler is the handler function for the UnIndexAll tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UnIndexAllHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/index/unindex", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UnIndexAll"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
