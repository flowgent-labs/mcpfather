package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreatePriorityScheme tool
const CreatePrioritySchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Data of priority scheme to create\",\n      \"properties\": {\n        \"defaultOptionId\": {\n          \"example\": \"3\",\n          \"type\": \"string\"\n        },\n        \"description\": {\n          \"example\": \"description\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": 10100,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"name\": {\n          \"example\": \"priority scheme name\",\n          \"type\": \"string\"\n        },\n        \"optionIds\": {\n          \"example\": [\n            \"1\",\n            \"2\",\n            \"3\",\n            \"4\",\n            \"5\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"1\\\",\\\"2\\\",\\\"3\\\",\\\"4\\\",\\\"5\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreatePriorityScheme tool (Status: 201, Content-Type: application/json)
const CreatePrioritySchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Newly created priority scheme\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n  - **optionIds** (Type: array):\n    - **Items** (Type: string):\n  - **projectKeys** (Type: array):\n    - **Items** (Type: string):\n  - **self** (Type: string, uri):\n  - **defaultOptionId** (Type: string):\n  - **defaultScheme** (Type: boolean):\n  - **description** (Type: string):\n  - **id** (Type: integer, int64):\n"

// NewCreatePrioritySchemeMCPTool creates the MCP Tool instance for CreatePriorityScheme
func NewCreatePrioritySchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreatePriorityScheme",
		"Create new priority scheme - Creates new priority scheme.",
		[]byte(CreatePrioritySchemeInputSchema),
	)
}

// CreatePrioritySchemeHandler is the handler function for the CreatePriorityScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreatePrioritySchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/priorityschemes", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreatePriorityScheme")
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
