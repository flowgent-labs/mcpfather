package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the AssignPriorityScheme tool
const AssignPrioritySchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Object that contains an id of the scheme\",\n      \"properties\": {\n        \"id\": {\n          \"example\": 10000,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"projectKeyOrId\": {\n      \"description\": \"Key or id of the project\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"projectKeyOrId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AssignPriorityScheme tool (Status: 200, Content-Type: application/json)
const AssignPrioritySchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Affected priority scheme.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n  - **optionIds** (Type: array):\n    - **Items** (Type: string):\n  - **projectKeys** (Type: array):\n    - **Items** (Type: string):\n  - **self** (Type: string, uri):\n  - **defaultOptionId** (Type: string):\n  - **defaultScheme** (Type: boolean):\n  - **description** (Type: string):\n  - **id** (Type: integer, int64):\n"

// NewAssignPrioritySchemeMCPTool creates the MCP Tool instance for AssignPriorityScheme
func NewAssignPrioritySchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AssignPriorityScheme",
		"Assign project with priority scheme - Assigns project with priority scheme. Priority scheme assign with migration is possible from the UI. Operation will fail if migration is needed as a result of operation eg. there are issues with priorities invalid in the destination scheme. All project keys associated with the priority scheme will only be returned if additional query parameter is provided expand=projectKeys.",
		[]byte(AssignPrioritySchemeInputSchema),
	)
}

// AssignPrioritySchemeHandler is the handler function for the AssignPriorityScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AssignPrioritySchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/project/{projectKeyOrId}/priorityscheme", args, []string{"projectKeyOrId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AssignPriorityScheme")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
