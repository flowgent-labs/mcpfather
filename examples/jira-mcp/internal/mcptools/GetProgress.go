package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetProgress tool
const GetProgressInputSchema = "{\n  \"properties\": {\n    \"requestId\": {\n      \"description\": \"The reindex request ID.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"requestId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProgress tool (Status: 200, Content-Type: application/json)
const GetProgressResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Details and status of the reindex request.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **type** (Type: string):\n      - Example: 'IMMEDIATE'\n      - Enum: ['IMMEDIATE', 'DELAYED']\n  - **completionTime** (Type: string, date-time):\n  - **id** (Type: integer, int64):\n      - Example: '10500'\n  - **requestTime** (Type: string, date-time):\n  - **startTime** (Type: string, date-time):\n  - **status** (Type: string):\n      - Example: 'PENDING'\n      - Enum: ['PENDING', 'ACTIVE', 'RUNNING', 'FAILED', 'COMPLETE']\n"

// NewGetProgressMCPTool creates the MCP Tool instance for GetProgress
func NewGetProgressMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProgress",
		"Get progress of a single reindex request - Retrieves the progress of a single reindex request.",
		[]byte(GetProgressInputSchema),
	)
}

// GetProgressHandler is the handler function for the GetProgress tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProgressHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/reindex/request/{requestId}", args, []string{"requestId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetProgress")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
