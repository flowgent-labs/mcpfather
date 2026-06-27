package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIdsOfWorklogsModifiedSince tool
const GetIdsOfWorklogsModifiedSinceInputSchema = "{\n  \"properties\": {\n    \"since\": {\n      \"default\": 0,\n      \"description\": \"a date time in unix timestamp format since when updated worklogs will be returned.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetIdsOfWorklogsModifiedSince tool (Status: 200, Content-Type: application/json)
const GetIdsOfWorklogsModifiedSinceResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the worklog changes.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/worklog/updated?since=1438013671136'\n  - **since** (Type: integer, int64):\n      - Example: '1438013671562'\n  - **until** (Type: integer, int64):\n      - Example: '1438013693136'\n  - **values** (Type: array):\n    - **Items** (Type: object):\n      - **worklogId** (Type: integer, int64):\n          - Example: '103'\n      - **updatedTime** (Type: integer, int64):\n          - Example: '1438013671562'\n  - **isLastPage** (Type: boolean):\n      - Example: 'true'\n  - **lastPage** (Type: boolean):\n  - **nextPage** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/worklog/updated?since=1438013693136'\n"

// NewGetIdsOfWorklogsModifiedSinceMCPTool creates the MCP Tool instance for GetIdsOfWorklogsModifiedSince
func NewGetIdsOfWorklogsModifiedSinceMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIdsOfWorklogsModifiedSince",
		"Returns worklogs updated since given time. - Returns worklogs id and update time of worklogs that was updated since given time. The returns set of worklogs is limited to 1000 elements. This API will not return worklogs updated during last minute.",
		[]byte(GetIdsOfWorklogsModifiedSinceInputSchema),
	)
}

// GetIdsOfWorklogsModifiedSinceHandler is the handler function for the GetIdsOfWorklogsModifiedSince tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIdsOfWorklogsModifiedSinceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/worklog/updated", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIdsOfWorklogsModifiedSince"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
