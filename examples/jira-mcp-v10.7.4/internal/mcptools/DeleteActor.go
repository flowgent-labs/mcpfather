package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DeleteActor tool
const DeleteActorInputSchema = "{\n  \"properties\": {\n    \"group\": {\n      \"description\": \"The group name to remove from the project role. Use either user or group, but not both\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"The project role id\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"The project id or project key\",\n      \"type\": \"string\"\n    },\n    \"user\": {\n      \"description\": \"The user name of the user to remove from the project role. Use either user or group, but not both\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteActorMCPTool creates the MCP Tool instance for DeleteActor
func NewDeleteActorMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteActor",
		"Delete actors from project role - Deletes actors (users or groups) from a project role.",
		[]byte(DeleteActorInputSchema),
	)
}

// DeleteActorHandler is the handler function for the DeleteActor tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteActorHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/project/{projectIdOrKey}/role/{id}", args, []string{"id", "projectIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteActor"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
