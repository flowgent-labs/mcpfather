package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAssignedPriorityScheme tool
const GetAssignedPrioritySchemeInputSchema = "{\n  \"properties\": {\n    \"projectKeyOrId\": {\n      \"description\": \"Key or id of the project\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectKeyOrId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAssignedPriorityScheme tool (Status: 200, Content-Type: application/json)
const GetAssignedPrioritySchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the priority scheme exists and the user has permission to view it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n  - **id** (Type: integer, int64):\n  - **name** (Type: string):\n  - **optionIds** (Type: array):\n    - **Items** (Type: string):\n  - **projectKeys** (Type: array):\n    - **Items** (Type: string):\n  - **self** (Type: string, uri):\n  - **defaultOptionId** (Type: string):\n  - **defaultScheme** (Type: boolean):\n"

// NewGetAssignedPrioritySchemeMCPTool creates the MCP Tool instance for GetAssignedPriorityScheme
func NewGetAssignedPrioritySchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAssignedPriorityScheme",
		"Get assigned priority scheme - Gets a full representation of a priority scheme in JSON format used by specified project. User must be global administrator or project administrator. All project keys associated with the priority scheme will only be returned if additional query parameter is provided expand=projectKeys.",
		[]byte(GetAssignedPrioritySchemeInputSchema),
	)
}

// GetAssignedPrioritySchemeHandler is the handler function for the GetAssignedPriorityScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAssignedPrioritySchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/project/{projectKeyOrId}/priorityscheme", args, []string{"projectKeyOrId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAssignedPriorityScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
