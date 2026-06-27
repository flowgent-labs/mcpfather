package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the RequestCurrentIndexFromNode tool
const RequestCurrentIndexFromNodeInputSchema = "{\n  \"properties\": {\n    \"nodeId\": {\n      \"description\": \"ID of the node to request index from\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"nodeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewRequestCurrentIndexFromNodeMCPTool creates the MCP Tool instance for RequestCurrentIndexFromNode
func NewRequestCurrentIndexFromNodeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RequestCurrentIndexFromNode",
		"Request node index snapshot - Request current index from node (the request is processed asynchronously). This method is deprecated as it is Lucene specific and is planned for removal in Jira 11.",
		[]byte(RequestCurrentIndexFromNodeInputSchema),
	)
}

// RequestCurrentIndexFromNodeHandler is the handler function for the RequestCurrentIndexFromNode tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RequestCurrentIndexFromNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/cluster/index-snapshot/{nodeId}", args, []string{"nodeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RequestCurrentIndexFromNode"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
