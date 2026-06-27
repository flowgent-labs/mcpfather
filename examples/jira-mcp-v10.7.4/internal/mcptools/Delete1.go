package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the Delete1 tool
const Delete1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"JSON containing parameters to replace the deleted version with\",\n      \"properties\": {\n        \"customFieldReplacementList\": {\n          \"items\": {\n            \"properties\": {\n              \"customFieldId\": {\n                \"example\": 2002,\n                \"format\": \"int64\",\n                \"type\": \"integer\"\n              },\n              \"moveTo\": {\n                \"example\": 10003,\n                \"format\": \"int64\",\n                \"type\": \"integer\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        },\n        \"moveAffectedIssuesTo\": {\n          \"example\": 10002,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"moveFixIssuesTo\": {\n          \"example\": 10001,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"The version to delete\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// NewDelete1MCPTool creates the MCP Tool instance for Delete1
func NewDelete1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Delete1",
		"Delete version and replace values - Delete a project version, removed values will be replaced with ones specified by the parameters.",
		[]byte(Delete1InputSchema),
	)
}

// Delete1Handler is the handler function for the Delete1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Delete1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/version/{id}/removeAndSwap", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Delete1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
