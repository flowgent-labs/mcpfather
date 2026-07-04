package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the Put2 tool
const Put2InputSchema = "{\n  \"properties\": {\n    \"If-Match\": {\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"the data to update the role with.\",\n      \"properties\": {\n        \"defaultGroups\": {\n          \"example\": [\n            \"jira-software-users\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"jira-software-users\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        },\n        \"defined\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"groups\": {\n          \"example\": [\n            \"jira-software-users\",\n            \"jira-testers\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"jira-software-users\\\",\\\"jira-testers\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        },\n        \"hasUnlimitedSeats\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"key\": {\n          \"example\": \"jira-software\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"Jira Software\",\n          \"type\": \"string\"\n        },\n        \"numberOfSeats\": {\n          \"example\": 10,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"platform\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"remainingSeats\": {\n          \"example\": 5,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"selectedByDefault\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"userCount\": {\n          \"example\": 5,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"userCountDescription\": {\n          \"example\": \"5 developers\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"key\": {\n      \"description\": \"the key of the role to update.\",\n      \"type\": \"string\"\n    },\n    \"versionHash\": {\n      \"description\": \"the hash of the version to update. Optional Param\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Put2 tool (Status: 200, Content-Type: application/json)
const Put2ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the updated ApplicationRole if the update was successful.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **userCount** (Type: integer, int32):\n      - Example: '5'\n  - **key** (Type: string):\n      - Example: 'jira-software'\n  - **numberOfSeats** (Type: integer, int32):\n      - Example: '10'\n  - **userCountDescription** (Type: string):\n      - Example: '5 developers'\n  - **hasUnlimitedSeats** (Type: boolean):\n      - Example: 'false'\n  - **name** (Type: string):\n      - Example: 'Jira Software'\n  - **platform** (Type: boolean):\n      - Example: 'false'\n  - **selectedByDefault** (Type: boolean):\n      - Example: 'false'\n  - **remainingSeats** (Type: integer, int32):\n      - Example: '5'\n  - **defaultGroups** (Type: array):\n      - Unique Items: true\n      - Example: '[\"jira-software-users\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-software-users\"]'\n  - **defined** (Type: boolean):\n      - Example: 'false'\n  - **groups** (Type: array):\n      - Unique Items: true\n      - Example: '[\"jira-software-users\",\"jira-testers\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-software-users\",\"jira-testers\"]'\n"

// NewPut2MCPTool creates the MCP Tool instance for Put2
func NewPut2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Put2",
		"Update application role - Updates the ApplicationRole with the passed data. Only the groups and default groups setting of the role may be updated. Requests to change the key or the name of the role will be silently ignored.",
		[]byte(Put2InputSchema),
	)
}

// Put2Handler is the handler function for the Put2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Put2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/applicationrole/{key}", args, []string{"key"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Put2")
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
