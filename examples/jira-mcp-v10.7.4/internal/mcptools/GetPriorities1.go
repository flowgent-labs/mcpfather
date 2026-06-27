package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPriorities1 tool
const GetPriorities1InputSchema = "{\n  \"properties\": {\n    \"maxResults\": {\n      \"default\": 100,\n      \"description\": \"how many results on the page should be included. Defaults to 100\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"projectIds\": {\n      \"description\": \"the list of project ids to filter priorities\",\n      \"items\": {\n        \"format\": \"int64\",\n        \"type\": \"integer\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"query\": {\n      \"default\": \"\",\n      \"description\": \"query that should match priority name or its translation\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"default\": 0,\n      \"description\": \"the page offset, if not specified then defaults to 0\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPriorities1 tool (Status: 200, Content-Type: application/json)
const GetPriorities1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> List of priorities\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'Major'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/priority/1'\n  - **statusColor** (Type: string):\n      - Example: 'red'\n  - **description** (Type: string):\n      - Example: 'This is a description of the priority'\n  - **iconUrl** (Type: string):\n      - Example: 'http://www.example.com/jira/images/icons/priorities/major.png'\n  - **id** (Type: string):\n      - Example: '1'\n"

// NewGetPriorities1MCPTool creates the MCP Tool instance for GetPriorities1
func NewGetPriorities1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPriorities1",
		"Get paginated issue priorities - Returns a page with list of issue priorities whose names (or their translations) match query",
		[]byte(GetPriorities1InputSchema),
	)
}

// GetPriorities1Handler is the handler function for the GetPriorities1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPriorities1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/priority/page", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPriorities1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
