package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPropertiesKeys4 tool
const GetPropertiesKeys4InputSchema = "{\n  \"properties\": {\n    \"userKey\": {\n      \"description\": \"Key of the user whose properties are to be returned\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"Username of the user whose properties are to be returned\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewGetPropertiesKeys4MCPTool creates the MCP Tool instance for GetPropertiesKeys4
func NewGetPropertiesKeys4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPropertiesKeys4",
		"Get keys of all properties for a user - Returns the keys of all properties for the user identified by the key or by the id.",
		[]byte(GetPropertiesKeys4InputSchema),
	)
}

// GetPropertiesKeys4Handler is the handler function for the GetPropertiesKeys4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPropertiesKeys4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/properties", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPropertiesKeys4"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
