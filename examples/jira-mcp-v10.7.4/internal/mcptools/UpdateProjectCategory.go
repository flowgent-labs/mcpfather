package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateProjectCategory tool
const UpdateProjectCategoryInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The project category to modify.\",\n      \"properties\": {\n        \"description\": {\n          \"example\": \"First Project Category\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"FIRST\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/projectCategory/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"Id of the project category to modify.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateProjectCategory tool (Status: 200, Content-Type: application/json)
const UpdateProjectCategoryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the project category exists and the currently authenticated user has permission to edit it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'My Project Category'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/projectCategory/10000'\n  - **description** (Type: string):\n      - Example: 'This is a project category'\n  - **id** (Type: string):\n      - Example: '10000'\n"

// NewUpdateProjectCategoryMCPTool creates the MCP Tool instance for UpdateProjectCategory
func NewUpdateProjectCategoryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateProjectCategory",
		"Update project category - Modify a project category.",
		[]byte(UpdateProjectCategoryInputSchema),
	)
}

// UpdateProjectCategoryHandler is the handler function for the UpdateProjectCategory tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateProjectCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/projectCategory/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateProjectCategory"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
