package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetTransitions tool
const GetTransitionsInputSchema = "{\n  \"properties\": {\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    },\n    \"transitionId\": {\n      \"description\": \"Transition id\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetTransitions tool (Status: 200, Content-Type: application/json)
const GetTransitionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a response containing a Map of TransitionFieldBeans for each transition possible by the current user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **transitions** (Type: array):\n    - **Items** (Type: object):\n      - **name** (Type: string):\n          - Example: 'Close Issue'\n      - **opsbarSequence** (Type: integer, int32):\n          - Example: '10'\n      - **to** (Type: object):\n        - **statusCategory** (Type: object):\n          - **self** (Type: string):\n              - Example: 'http://localhost:8090/jira/rest/api/2.0/statuscategory/1'\n          - **colorName** (Type: string):\n              - Example: 'blue-gray'\n          - **id** (Type: integer, int64):\n              - Example: '1'\n          - **key** (Type: string):\n              - Example: 'new'\n          - **name** (Type: string):\n              - Example: 'To Do'\n        - **statusColor** (Type: string):\n            - Example: 'green'\n        - **description** (Type: string):\n            - Example: 'The issue is currently being worked on.'\n        - **iconUrl** (Type: string):\n            - Example: 'http://localhost:8090/jira/images/icons/progress.gif'\n        - **id** (Type: string):\n            - Example: '10000'\n        - **name** (Type: string):\n            - Example: 'In Progress'\n        - **self** (Type: string):\n            - Example: 'http://localhost:8090/jira/rest/api/2.0/status/10000'\n      - **description** (Type: string):\n          - Example: 'Close the issue.'\n      - **fields** (Type: object):\n        - **Additional Properties**:\n          - **property value** (Type: object):\n            - **allowedValues** (Type: array):\n                - Example: '[\"red\",\"blue\",\"default value\"]'\n              - **Items** (Type: object):\n                  - Example: '[\"red\",\"blue\",\"default value\"]'\n            - **autoCompleteUrl** (Type: string):\n                - Example: '/rest/api/2/customFieldOption/10000'\n            - **fieldId** (Type: string):\n                - Example: 'customfield_10000'\n            - **required** (Type: boolean):\n                - Example: 'true'\n            - **schema** (Type: object):\n                - Example: '{}'\n              - **items** (Type: string):\n                  - Example: 'null'\n              - **system** (Type: string):\n                  - Example: 'summary'\n              - **type** (Type: string):\n                  - Example: 'string'\n              - **custom** (Type: string):\n                  - Example: 'null'\n              - **customId** (Type: integer, int64):\n            - **defaultValue** (Type: object):\n            - **name** (Type: string):\n                - Example: 'My Multi Select'\n            - **hasDefaultValue** (Type: boolean):\n                - Example: 'true'\n            - **operations** (Type: array):\n                - Example: '[\"set\",\"add\"]'\n              - **Items** (Type: string):\n                  - Example: '[\"set\",\"add\"]'\n      - **id** (Type: string):\n          - Example: '2'\n"

// NewGetTransitionsMCPTool creates the MCP Tool instance for GetTransitions
func NewGetTransitionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetTransitions",
		"Get list of transitions possible for an issue - Get a list of the transitions possible for this issue by the current user, along with fields that are required and their types.\nFields will only be returned if "+"\x60"+"expand=transitions.fields"+"\x60"+".\nThe fields in the metadata correspond to the fields in the transition screen for that transition.\nFields not in the screen will not be in the metadata.",
		[]byte(GetTransitionsInputSchema),
	)
}

// GetTransitionsHandler is the handler function for the GetTransitions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetTransitionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issue/{issueIdOrKey}/transitions", args, []string{"issueIdOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetTransitions")
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
