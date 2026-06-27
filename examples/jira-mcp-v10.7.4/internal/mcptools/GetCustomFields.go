package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetCustomFields tool
const GetCustomFieldsInputSchema = "{\n  \"properties\": {\n    \"lastValueUpdate\": {\n      \"description\": \"The last value update to filter the custom fields.\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of custom fields to return.\",\n      \"type\": \"string\"\n    },\n    \"projectIds\": {\n      \"description\": \"A list of project IDs to filter the custom fields.\",\n      \"type\": \"string\"\n    },\n    \"screenIds\": {\n      \"description\": \"A list of screen IDs to filter the custom fields.\",\n      \"type\": \"string\"\n    },\n    \"search\": {\n      \"description\": \"A query string used to search custom fields.\",\n      \"type\": \"string\"\n    },\n    \"sortColumn\": {\n      \"description\": \"The column by which to sort the returned custom fields.\",\n      \"type\": \"string\"\n    },\n    \"sortOrder\": {\n      \"description\": \"The order in which to sort the returned custom fields.\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"description\": \"The starting index of the returned custom fields.\",\n      \"type\": \"string\"\n    },\n    \"types\": {\n      \"description\": \"A list of custom field types to filter the custom fields.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetCustomFields tool (Status: 200, Content-Type: application/json)
const GetCustomFieldsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if a custom field with the given customFieldId exists and user has permission to it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **numericId** (Type: integer, int64):\n      - Example: '10000'\n  - **lastValueUpdate** (Type: string, date-time):\n      - Example: '2018-11-01T12:00:00Z'\n  - **name** (Type: string):\n      - Example: 'New custom field'\n  - **screensCount** (Type: integer, int32):\n      - Example: '3'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **isTrusted** (Type: boolean):\n      - Example: 'true'\n  - **description** (Type: string):\n      - Example: 'Custom field for picking groups'\n  - **issueTypeIds** (Type: array):\n      - Example: '[\"1\",\"2\"]'\n    - **Items** (Type: string):\n        - Example: '[\"1\",\"2\"]'\n  - **searcherKey** (Type: string):\n      - Example: 'com.atlassian.jira.plugin.system.customfieldtypes:grouppickersearcher'\n  - **isAllProjects** (Type: boolean):\n      - Example: 'false'\n  - **self** (Type: string, uri):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/customField/10000'\n  - **issuesWithValue** (Type: integer, int64):\n      - Example: '100'\n  - **isManaged** (Type: boolean):\n      - Example: 'false'\n  - **type** (Type: string):\n      - Example: 'com.atlassian.jira.plugin.system.customfieldtypes:grouppicker'\n  - **projectIds** (Type: array):\n      - Example: '[10000,10001]'\n    - **Items** (Type: integer, int64):\n  - **isLocked** (Type: boolean):\n      - Example: 'false'\n  - **projectsCount** (Type: integer, int32):\n      - Example: '2'\n"

// NewGetCustomFieldsMCPTool creates the MCP Tool instance for GetCustomFields
func NewGetCustomFieldsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCustomFields",
		"Get custom fields with pagination - Returns a list of Custom Fields in the given range.",
		[]byte(GetCustomFieldsInputSchema),
	)
}

// GetCustomFieldsHandler is the handler function for the GetCustomFields tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCustomFieldsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/customFields", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetCustomFields"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
