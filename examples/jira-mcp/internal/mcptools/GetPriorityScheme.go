package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPriorityScheme tool
const GetPrioritySchemeInputSchema = "{\n  \"properties\": {\n    \"schemeId\": {\n      \"description\": \"id of priority scheme to get\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPriorityScheme tool (Status: 200, Content-Type: application/json)
const GetPrioritySchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Priority scheme\n\n## Response Structure\n\n- Structure (Type: object):\n  - **defaultOptionId** (Type: string):\n  - **defaultScheme** (Type: boolean):\n  - **description** (Type: string):\n  - **id** (Type: integer, int64):\n  - **name** (Type: string):\n  - **optionIds** (Type: array):\n    - **Items** (Type: string):\n  - **projectKeys** (Type: array):\n    - **Items** (Type: string):\n  - **self** (Type: string, uri):\n"

// NewGetPrioritySchemeMCPTool creates the MCP Tool instance for GetPriorityScheme
func NewGetPrioritySchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPriorityScheme",
		"Get a priority scheme by ID - Gets a full representation of a priority scheme in JSON format.",
		[]byte(GetPrioritySchemeInputSchema),
	)
}

// GetPrioritySchemeHandler is the handler function for the GetPriorityScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPrioritySchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/priorityschemes/{schemeId}", args, []string{"schemeId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPriorityScheme")
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
