package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetCustomFieldOptions tool
const GetCustomFieldOptionsInputSchema = "{\n  \"properties\": {\n    \"customFieldId\": {\n      \"description\": \"The ID of the custom field.\",\n      \"type\": \"string\"\n    },\n    \"issueTypeIds\": {\n      \"description\": \"A list of issue type IDs in a context.\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of results to return.\",\n      \"type\": \"string\"\n    },\n    \"page\": {\n      \"description\": \"The page of options to return.\",\n      \"type\": \"string\"\n    },\n    \"projectIds\": {\n      \"description\": \"A list of project IDs in a context.\",\n      \"type\": \"string\"\n    },\n    \"query\": {\n      \"description\": \"A string used to filter options.\",\n      \"type\": \"string\"\n    },\n    \"sortByOptionName\": {\n      \"description\": \"Flag to sort options by their names.\",\n      \"type\": \"string\"\n    },\n    \"useAllContexts\": {\n      \"description\": \"Flag to fetch all options regardless of context, project IDs, or issue type IDs.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"customFieldId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetCustomFieldOptions tool (Status: 200, Content-Type: application/json)
const GetCustomFieldOptionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if a custom field with the given customFieldId exists and user has permission to it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **options** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: integer, int64):\n          - Example: '3'\n      - **self** (Type: string, uri):\n          - Example: 'http://localhost:8090/jira/rest/api/2.0/customFieldOption/3'\n      - **value** (Type: string):\n          - Example: 'Blue'\n      - **childrenIds** (Type: array):\n          - Example: '[4,5]'\n        - **Items** (Type: integer, int64):\n      - **disabled** (Type: boolean):\n          - Example: 'false'\n  - **total** (Type: integer, int32):\n      - Example: '1'\n"

// NewGetCustomFieldOptionsMCPTool creates the MCP Tool instance for GetCustomFieldOptions
func NewGetCustomFieldOptionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCustomFieldOptions",
		"Get custom field options - Returns custom field's options defined in a given context composed of projects and issue types.",
		[]byte(GetCustomFieldOptionsInputSchema),
	)
}

// GetCustomFieldOptionsHandler is the handler function for the GetCustomFieldOptions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCustomFieldOptionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/customFields/{customFieldId}/options", args, []string{"customFieldId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetCustomFieldOptions"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
