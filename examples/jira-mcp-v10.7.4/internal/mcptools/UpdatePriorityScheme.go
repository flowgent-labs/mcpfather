package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdatePriorityScheme tool
const UpdatePrioritySchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"New scheme data\",\n      \"properties\": {\n        \"defaultOptionId\": {\n          \"example\": \"3\",\n          \"type\": \"string\"\n        },\n        \"description\": {\n          \"example\": \"description\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": 10100,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"name\": {\n          \"example\": \"priority scheme name\",\n          \"type\": \"string\"\n        },\n        \"optionIds\": {\n          \"example\": [\n            \"1\",\n            \"2\",\n            \"3\",\n            \"4\",\n            \"5\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"1\\\",\\\"2\\\",\\\"3\\\",\\\"4\\\",\\\"5\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"schemeId\": {\n      \"description\": \"id of the priority scheme to update\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdatePriorityScheme tool (Status: 200, Content-Type: application/json)
const UpdatePrioritySchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Updated priority scheme\n\n## Response Structure\n\n- Structure (Type: object):\n  - **defaultScheme** (Type: boolean):\n  - **description** (Type: string):\n  - **id** (Type: integer, int64):\n  - **name** (Type: string):\n  - **optionIds** (Type: array):\n    - **Items** (Type: string):\n  - **projectKeys** (Type: array):\n    - **Items** (Type: string):\n  - **self** (Type: string, uri):\n  - **defaultOptionId** (Type: string):\n"

// NewUpdatePrioritySchemeMCPTool creates the MCP Tool instance for UpdatePriorityScheme
func NewUpdatePrioritySchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdatePriorityScheme",
		"Update a priority scheme - Updates a priority scheme. Update will be rejected if issue migration would be needed as a result of scheme update. Priority scheme update with migration is possible from the UI.",
		[]byte(UpdatePrioritySchemeInputSchema),
	)
}

// UpdatePrioritySchemeHandler is the handler function for the UpdatePriorityScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdatePrioritySchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/priorityschemes/{schemeId}", args, []string{"schemeId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdatePriorityScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
