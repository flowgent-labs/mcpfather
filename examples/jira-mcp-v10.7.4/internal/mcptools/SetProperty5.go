package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetProperty5 tool
const SetProperty5InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request containing value of the user's property. The value has to be a valid, non-empty JSON conforming to http://tools.ietf.org/html/rfc4627. The maximum length of the property value is 32768 bytes.\",\n      \"type\": \"string\"\n    },\n    \"propertyKey\": {\n      \"description\": \"The key of the user's property. The maximum length of the key is 255 bytes.\",\n      \"type\": \"string\"\n    },\n    \"userKey\": {\n      \"description\": \"Key of the user whose property is to be set\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"Username of the user whose property is to be set\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"propertyKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProperty5MCPTool creates the MCP Tool instance for SetProperty5
func NewSetProperty5MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProperty5",
		"Set the value of a specified user's property - Sets the value of the specified user's property.\nYou can use this resource to store a custom data against the user identified by the key or by the id. The user\nwho stores the data is required to have permissions to administer the user.",
		[]byte(SetProperty5InputSchema),
	)
}

// SetProperty5Handler is the handler function for the SetProperty5 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProperty5Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/user/properties/{propertyKey}", args, []string{"propertyKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetProperty5"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
