package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateIssueTypeScheme tool
const CreateIssueTypeSchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Issue type scheme creation details.\",\n      \"properties\": {\n        \"defaultIssueTypeId\": {\n          \"example\": \"3\",\n          \"type\": \"string\"\n        },\n        \"description\": {\n          \"example\": \"some new description of the scheme\",\n          \"type\": \"string\"\n        },\n        \"issueTypeIDs\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"writeOnly\": true\n        },\n        \"issueTypeIds\": {\n          \"example\": [\n            \"1\",\n            \"4\",\n            \"3\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"1\\\",\\\"4\\\",\\\"3\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"name\": {\n          \"example\": \"new name\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateIssueTypeScheme tool (Status: 200, Content-Type: application/json)
const CreateIssueTypeSchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the newly created IssueTypeScheme if successful.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **expand** (Type: string):\n      - Example: 'issueTypes'\n  - **id** (Type: string):\n      - Example: '10100'\n  - **issueTypes** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n          - Example: '1'\n      - **name** (Type: string):\n          - Example: 'Bug'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n      - **subtask** (Type: boolean):\n          - Example: 'false'\n      - **avatarId** (Type: integer, int64):\n          - Example: '10002'\n      - **description** (Type: string):\n          - Example: 'A problem which impairs or prevents the functions of the product.'\n      - **iconUrl** (Type: string):\n          - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n  - **name** (Type: string):\n      - Example: 'Some grouping of issue types for the greater good.'\n  - **self** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2/issuetypescheme/10100'\n  - **[cyclic reference]**\n  - **description** (Type: string):\n      - Example: 'A brief explanation of this issue type scheme.'\n"

// NewCreateIssueTypeSchemeMCPTool creates the MCP Tool instance for CreateIssueTypeScheme
func NewCreateIssueTypeSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateIssueTypeScheme",
		"Create an issue type scheme from JSON representation - Creates an issue type scheme from a JSON representation",
		[]byte(CreateIssueTypeSchemeInputSchema),
	)
}

// CreateIssueTypeSchemeHandler is the handler function for the CreateIssueTypeScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateIssueTypeSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issuetypescheme", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateIssueTypeScheme")
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
