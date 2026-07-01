package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateProjectCategory tool
const CreateProjectCategoryInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The project category to create.\",\n      \"properties\": {\n        \"description\": {\n          \"example\": \"First Project Category\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"FIRST\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/projectCategory/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateProjectCategory tool (Status: 201, Content-Type: application/json)
const CreateProjectCategoryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returned if the project category is created successfully.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'My Project Category'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/projectCategory/10000'\n  - **description** (Type: string):\n      - Example: 'This is a project category'\n  - **id** (Type: string):\n      - Example: '10000'\n"

// NewCreateProjectCategoryMCPTool creates the MCP Tool instance for CreateProjectCategory
func NewCreateProjectCategoryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateProjectCategory",
		"Create project category - Create a project category.",
		[]byte(CreateProjectCategoryInputSchema),
	)
}

// CreateProjectCategoryHandler is the handler function for the CreateProjectCategory tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateProjectCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/projectCategory", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateProjectCategory")
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
