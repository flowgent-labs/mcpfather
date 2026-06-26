package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SearchForProjects tool
const SearchForProjectsInputSchema = "{\n  \"properties\": {\n    \"allowEmptyQuery\": {\n      \"default\": false,\n      \"description\": \"If true, and the query is empty, the method will return first results limited to the value of 'maxResults' or default limit of 100.\",\n      \"type\": \"boolean\"\n    },\n    \"maxResults\": {\n      \"default\": 0,\n      \"description\": \"Maximum number of matches to return. Zero means a default limit of 100 and negative numbers return no results.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"query\": {\n      \"default\": \"\",\n      \"description\": \"A sequence of characters expected to be found in the word-prefix of project name and/or key.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the SearchForProjects tool (Status: 200, Content-Type: application/json)
const SearchForProjectsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned even when no projects match the given query.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **header** (Type: string):\n      - Example: 'Showing 2 of 5 matching projects'\n  - **projects** (Type: array):\n    - **Items** (Type: object):\n      - **name** (Type: string):\n          - Example: 'Example project name'\n      - **avatar** (Type: string):\n          - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n      - **html** (Type: string):\n          - Example: 'Example <strong>pro</strong>ject name (EXAM)'\n      - **id** (Type: string):\n          - Example: '10000'\n      - **key** (Type: string):\n          - Example: 'EXAM'\n  - **total** (Type: integer, int32):\n      - Example: '5'\n"

// NewSearchForProjectsMCPTool creates the MCP Tool instance for SearchForProjects
func NewSearchForProjectsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SearchForProjects",
		"Get projects matching query - Returns a list of projects visible to the user where project name and/or key is matching the given query.\nPassing an empty (or whitespace only) query will match no projects. The project matches will\ncontain a field with the query highlighted.\nThe number of projects returned can be controlled by passing a value for 'maxResults', but a hard limit of no\nmore than 100 projects is enforced. The projects are wrapped in a single response object that contains\na header for use in the picker, specifically 'Showing X of Y matching projects' and the total number\nof matches for the query.",
		[]byte(SearchForProjectsInputSchema),
	)
}

// SearchForProjectsHandler is the handler function for the SearchForProjects tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SearchForProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/projects/picker", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SearchForProjects"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
