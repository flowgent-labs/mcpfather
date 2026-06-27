package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProgressBulk tool
const GetProgressBulkInputSchema = "{\n  \"properties\": {\n    \"requestId\": {\n      \"description\": \"The reindex request IDs.\",\n      \"items\": {\n        \"format\": \"int64\",\n        \"type\": \"integer\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetProgressBulk tool (Status: 200, Content-Type: application/json)
const GetProgressBulkResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> An array of results describing the progress of each of the found requests.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **requestTime** (Type: string, date-time):\n  - **startTime** (Type: string, date-time):\n  - **status** (Type: string):\n      - Example: 'PENDING'\n      - Enum: ['PENDING', 'ACTIVE', 'RUNNING', 'FAILED', 'COMPLETE']\n  - **type** (Type: string):\n      - Example: 'IMMEDIATE'\n      - Enum: ['IMMEDIATE', 'DELAYED']\n  - **completionTime** (Type: string, date-time):\n  - **id** (Type: integer, int64):\n      - Example: '10500'\n"

// NewGetProgressBulkMCPTool creates the MCP Tool instance for GetProgressBulk
func NewGetProgressBulkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProgressBulk",
		"Get progress of multiple reindex requests - Retrieves the progress of multiple reindex requests. Only reindex requests that actually exist will be returned in the results.",
		[]byte(GetProgressBulkInputSchema),
	)
}

// GetProgressBulkHandler is the handler function for the GetProgressBulk tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProgressBulkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/reindex/request/bulk", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProgressBulk"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
