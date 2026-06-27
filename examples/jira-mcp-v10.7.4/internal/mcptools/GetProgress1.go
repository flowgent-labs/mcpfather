package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProgress1 tool
const GetProgress1InputSchema = "{\n  \"properties\": {\n    \"taskId\": {\n      \"description\": \"The id of a user anonymization task you wish to obtain details on.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewGetProgress1MCPTool creates the MCP Tool instance for GetProgress1
func NewGetProgress1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProgress1",
		"Get user anonymization progress - Returns information about a user anonymization operation progress.",
		[]byte(GetProgress1InputSchema),
	)
}

// GetProgress1Handler is the handler function for the GetProgress1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProgress1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/anonymization/progress", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProgress1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
