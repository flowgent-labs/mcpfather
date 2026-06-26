package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetProperty3 tool
const SetProperty3InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The value of the issue type's property. The value has to be a valid, non-empty JSON conforming to http://tools.ietf.org/html/rfc4627. The maximum length of the property value is 32768 bytes.\",\n      \"properties\": {\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"key\": {\n          \"type\": \"string\"\n        },\n        \"value\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueTypeId\": {\n      \"description\": \"The issue type on which the property will be set.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the issue type's property. The maximum length of the key is 255 bytes\",\n      \"maxLength\": 255,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"issueTypeId\",\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty3MCPTool creates the MCP Tool instance for SetProperty3
func NewSetProperty3MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty3",
		"Update specified issue type's property - Sets the value of the specified issue type's property",
		[]byte(SetProperty3InputSchema),
	)
}

// SetProperty3Handler is the handler function for the SetProperty3 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProperty3Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issuetype/{issueTypeId}/properties/{propertyKey}", args, []string{"issueTypeId", "propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetProperty3"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
