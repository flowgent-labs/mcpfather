package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateCustomField tool
const CreateCustomFieldInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"description\": {\n          \"example\": \"Custom field for picking groups\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"issueTypeIds\": {\n          \"example\": [\n            \"1\",\n            \"2\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"1\\\",\\\"2\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"name\": {\n          \"example\": \"New custom field\",\n          \"type\": \"string\"\n        },\n        \"projectIds\": {\n          \"example\": [\n            10000,\n            10001\n          ],\n          \"items\": {\n            \"format\": \"int64\",\n            \"type\": \"integer\"\n          },\n          \"type\": \"array\"\n        },\n        \"searcherKey\": {\n          \"example\": \"com.atlassian.jira.plugin.system.customfieldtypes:grouppickersearcher\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://localhost:8090/jira/rest/api/2.0/customField/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"type\": {\n          \"example\": \"com.atlassian.jira.plugin.system.customfieldtypes:grouppicker\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the CreateCustomField tool (Status: 201, Content-Type: application/json)
const CreateCustomFieldResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Custom field was created\n\n## Response Structure\n\n- Structure (Type: object):\n  - **clauseNames** (Type: array):\n      - Unique Items: true\n      - Example: '\"[description]\"'\n    - **Items** (Type: string):\n        - Example: '[description]'\n  - **custom** (Type: boolean):\n      - Example: 'false'\n  - **id** (Type: string):\n      - Example: 'description'\n  - **name** (Type: string):\n      - Example: 'Description'\n  - **navigable** (Type: boolean):\n      - Example: 'true'\n  - **orderable** (Type: boolean):\n      - Example: 'true'\n  - **schema** (Type: object):\n      - Example: '{}'\n    - **custom** (Type: string):\n        - Example: 'null'\n    - **customId** (Type: integer, int64):\n    - **items** (Type: string):\n        - Example: 'null'\n    - **system** (Type: string):\n        - Example: 'summary'\n    - **type** (Type: string):\n        - Example: 'string'\n  - **searchable** (Type: boolean):\n      - Example: 'true'\n"

// NewCreateCustomFieldMCPTool creates the MCP Tool instance for CreateCustomField
func NewCreateCustomFieldMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateCustomField",
		"Create a custom field using a definition - Creates a custom field using a definition",
		[]byte(CreateCustomFieldInputSchema),
	)
}

// CreateCustomFieldHandler is the handler function for the CreateCustomField tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateCustomFieldHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/field", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateCustomField")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
