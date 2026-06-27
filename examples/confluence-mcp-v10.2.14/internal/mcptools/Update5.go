package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Update5 tool
const Update5InputSchema = "{\n  \"properties\": {\n    \"groupName\": {\n      \"description\": \"The group name identifying the given group.\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"The username identifying the given user.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupName\",\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdate5MCPTool creates the MCP Tool instance for Update5
func NewUpdate5MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Update5",
		"Update user group - Add the given User identified by username to the given Group identified by groupName. \n\nThis method is idempotent i.e., if the membership already exists then no action will be taken.",
		[]byte(Update5InputSchema),
	)
}

// Update5Handler is the handler function for the Update5 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Update5Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/user/{username}/group/{groupName}", args, []string{"groupName", "username"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Update5"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
