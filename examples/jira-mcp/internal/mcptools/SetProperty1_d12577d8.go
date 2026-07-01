package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the SetProperty1_d12577d8 tool
const SetProperty1_d12577d8InputSchema = "{\n  \"properties\": {\n    \"propertyKey\": {\n      \"description\": \"The key of the sprint's property. The maximum length of the key is 255 bytes.\",\n      \"type\": \"string\"\n    },\n    \"sprintId\": {\n      \"description\": \"The id of the sprint on which the property will be set.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"propertyKey\",\n    \"sprintId\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty1_d12577d8MCPTool creates the MCP Tool instance for SetProperty1_d12577d8
func NewSetProperty1_d12577d8MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty1_d12577d8",
		"Update a sprint's property - Sets the value of the specified sprint's property. You can use this resource to store a custom data against the sprint identified by the id. The user who stores the data is required to have permissions to modify the sprint.",
		[]byte(SetProperty1_d12577d8InputSchema),
	)
}

// SetProperty1_d12577d8Handler is the handler function for the SetProperty1_d12577d8 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProperty1_d12577d8Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/agile/1.0/sprint/{sprintId}/properties/{propertyKey}", args, []string{"propertyKey", "sprintId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetProperty1_d12577d8")
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
