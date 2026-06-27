package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProjectCategoryById tool
const GetProjectCategoryByIdInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"A project category id\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProjectCategoryById tool (Status: 200, Content-Type: application/json)
const GetProjectCategoryByIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the project category exists and is visible by the calling user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n      - Example: 'This is a project category'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **name** (Type: string):\n      - Example: 'My Project Category'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/projectCategory/10000'\n"

// NewGetProjectCategoryByIdMCPTool creates the MCP Tool instance for GetProjectCategoryById
func NewGetProjectCategoryByIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProjectCategoryById",
		"Get project category by ID - Returns a full representation of the project category that has the given id.",
		[]byte(GetProjectCategoryByIdInputSchema),
	)
}

// GetProjectCategoryByIdHandler is the handler function for the GetProjectCategoryById tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProjectCategoryByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/projectCategory/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProjectCategoryById"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
