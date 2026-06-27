package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UnassignPriorityScheme tool
const UnassignPrioritySchemeInputSchema = "{\n  \"properties\": {\n    \"projectKeyOrId\": {\n      \"description\": \"Key or id of the project\",\n      \"type\": \"string\"\n    },\n    \"schemeId\": {\n      \"description\": \"Object that contains an id of the scheme\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"projectKeyOrId\",\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UnassignPriorityScheme tool (Status: 200, Content-Type: application/json)
const UnassignPrioritySchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Affected priority scheme.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n  - **optionIds** (Type: array):\n    - **Items** (Type: string):\n  - **projectKeys** (Type: array):\n    - **Items** (Type: string):\n  - **self** (Type: string, uri):\n  - **defaultOptionId** (Type: string):\n  - **defaultScheme** (Type: boolean):\n  - **description** (Type: string):\n  - **id** (Type: integer, int64):\n"

// NewUnassignPrioritySchemeMCPTool creates the MCP Tool instance for UnassignPriorityScheme
func NewUnassignPrioritySchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UnassignPriorityScheme",
		"Unassign project from priority scheme - Unassigns project from priority scheme. Operation will fail for defualt priority scheme, project is not found or project is not associated with provided priority scheme. All project keys associated with the priority scheme will only be returned if additional query parameter is provided expand=projectKeys.",
		[]byte(UnassignPrioritySchemeInputSchema),
	)
}

// UnassignPrioritySchemeHandler is the handler function for the UnassignPriorityScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UnassignPrioritySchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/project/{projectKeyOrId}/priorityscheme/{schemeId}", args, []string{"projectKeyOrId", "schemeId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UnassignPriorityScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
