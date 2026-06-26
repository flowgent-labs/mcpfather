package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAssociatedProjects tool
const GetAssociatedProjectsInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"type\": \"string\"\n    },\n    \"schemeId\": {\n      \"description\": \"Id of the issue type scheme whose projects we're accessing\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAssociatedProjects tool (Status: 200, Content-Type: application/json)
const GetAssociatedProjectsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The collection of projects associated with this issue type scheme.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n      - Example: 'Example'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **key** (Type: string):\n      - Example: 'EX'\n  - **name** (Type: string):\n      - Example: 'Example'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/project/EX'\n  - **archived** (Type: boolean):\n      - Example: 'false'\n  - **avatarUrls** (Type: object):\n      - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n"

// NewGetAssociatedProjectsMCPTool creates the MCP Tool instance for GetAssociatedProjects
func NewGetAssociatedProjectsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAssociatedProjects",
		"Get all of the associated projects for specified scheme - For the specified issue type scheme, returns all of the associated projects",
		[]byte(GetAssociatedProjectsInputSchema),
	)
}

// GetAssociatedProjectsHandler is the handler function for the GetAssociatedProjects tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAssociatedProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuetypescheme/{schemeId}/associations", args, []string{"schemeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAssociatedProjects"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
