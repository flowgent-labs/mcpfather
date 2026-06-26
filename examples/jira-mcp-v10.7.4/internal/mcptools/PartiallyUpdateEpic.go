package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the PartiallyUpdateEpic tool
const PartiallyUpdateEpicInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The epic properties to update.\",\n      \"properties\": {\n        \"color\": {\n          \"example\": \"color_6\",\n          \"properties\": {\n            \"key\": {\n              \"enum\": [\n                \"color_1\",\n                \"color_2\",\n                \"color_3\",\n                \"color_4\",\n                \"color_5\",\n                \"color_6\",\n                \"color_7\",\n                \"color_8\",\n                \"color_9\",\n                \"color_10\",\n                \"color_11\",\n                \"color_12\",\n                \"color_13\",\n                \"color_14\"\n              ],\n              \"example\": \"ghx-label-1\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"done\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"name\": {\n          \"example\": \"Epic 1\",\n          \"type\": \"string\"\n        },\n        \"summary\": {\n          \"example\": \"Epic 1 summary\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"epicIdOrKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"epicIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PartiallyUpdateEpic tool (Status: 200, Content-Type: application/json)
const PartiallyUpdateEpicResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Updated epic\n\n## Response Structure\n\n- Structure (Type: object):\n  - **color** (Type: object):\n      - Example: '\"color_6\"'\n    - **key** (Type: string):\n        - Example: 'ghx-label-1'\n        - Enum: ['color_1', 'color_2', 'color_3', 'color_4', 'color_5', 'color_6', 'color_7', 'color_8', 'color_9', 'color_10', 'color_11', 'color_12', 'color_13', 'color_14']\n  - **done** (Type: boolean):\n      - Example: 'true'\n  - **id** (Type: integer, int64):\n      - Example: '10000'\n  - **key** (Type: string):\n      - Example: 'PR-1'\n  - **name** (Type: string):\n      - Example: 'Epic 1'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issue/10000'\n  - **summary** (Type: string):\n      - Example: 'Epic 1 summary'\n"

// NewPartiallyUpdateEpicMCPTool creates the MCP Tool instance for PartiallyUpdateEpic
func NewPartiallyUpdateEpicMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PartiallyUpdateEpic",
		"Update an epic's details - Performs a partial update of the epic. A partial update means that fields not present in the request JSON will not be updated. Valid values for color are color_1 to color_9.",
		[]byte(PartiallyUpdateEpicInputSchema),
	)
}

// PartiallyUpdateEpicHandler is the handler function for the PartiallyUpdateEpic tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PartiallyUpdateEpicHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/epic/{epicIdOrKey}", args, []string{"epicIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "PartiallyUpdateEpic"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
