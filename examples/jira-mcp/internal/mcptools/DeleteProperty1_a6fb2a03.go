package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the DeleteProperty1_a6fb2a03 tool
const DeleteProperty1_a6fb2a03InputSchema = "{\n  \"properties\": {\n    \"propertyKey\": {\n      \"description\": \"The key of the property to remove.\",\n      \"type\": \"string\"\n    },\n    \"sprintId\": {\n      \"description\": \"The id of the sprint from which the property will be removed.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"propertyKey\",\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteProperty1_a6fb2a03MCPTool creates the MCP Tool instance for DeleteProperty1_a6fb2a03
func NewDeleteProperty1_a6fb2a03MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteProperty1_a6fb2a03",
		"Delete a sprint's property - Removes the property from the sprint identified by the id. Ths user removing the property is required to have permissions to modify the sprint.",
		[]byte(DeleteProperty1_a6fb2a03InputSchema),
	)
}

// DeleteProperty1_a6fb2a03Handler is the handler function for the DeleteProperty1_a6fb2a03 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteProperty1_a6fb2a03Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/agile/1.0/sprint/{sprintId}/properties/{propertyKey}", args, []string{"propertyKey", "sprintId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteProperty1_a6fb2a03")
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
