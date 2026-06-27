package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPaginatedStatuses tool
const GetPaginatedStatusesInputSchema = "{\n  \"properties\": {\n    \"issueTypeIds\": {\n      \"description\": \"The list of issue type ids to filter statuses.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"maxResults\": {\n      \"default\": 100,\n      \"description\": \"The maximum number of statuses to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"projectIds\": {\n      \"description\": \"The list of project ids to filter statuses.\",\n      \"items\": {\n        \"format\": \"int64\",\n        \"type\": \"integer\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"query\": {\n      \"default\": \"\",\n      \"description\": \"The string that status names will be matched with.\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"default\": 0,\n      \"description\": \"The index of the first status to return.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPaginatedStatuses tool (Status: 200, Content-Type: application/json)
const GetPaginatedStatusesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns paginated list of statuses.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **iconUrl** (Type: string):\n      - Example: 'http://localhost:8090/jira/images/icons/progress.gif'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **name** (Type: string):\n      - Example: 'In Progress'\n  - **self** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/status/10000'\n  - **statusCategory** (Type: object):\n    - **key** (Type: string):\n        - Example: 'new'\n    - **name** (Type: string):\n        - Example: 'To Do'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8090/jira/rest/api/2.0/statuscategory/1'\n    - **colorName** (Type: string):\n        - Example: 'blue-gray'\n    - **id** (Type: integer, int64):\n        - Example: '1'\n  - **statusColor** (Type: string):\n      - Example: 'green'\n  - **description** (Type: string):\n      - Example: 'The issue is currently being worked on.'\n"

// NewGetPaginatedStatusesMCPTool creates the MCP Tool instance for GetPaginatedStatuses
func NewGetPaginatedStatusesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPaginatedStatuses",
		"Get paginated filtered statuses - Returns paginated list of filtered statuses",
		[]byte(GetPaginatedStatusesInputSchema),
	)
}

// GetPaginatedStatusesHandler is the handler function for the GetPaginatedStatuses tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPaginatedStatusesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/status/page", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPaginatedStatuses"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
