package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetProjects tool
const GetProjectsInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"description\": \"The Id of the board that contains returned projects.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of projects to return per page. Default: 50. See the 'Pagination' section at the top of this page for more details.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"startAt\": {\n      \"description\": \"The starting index of the returned projects. Base index: 0. See the 'Pagination' section at the top of this page for more details.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"boardId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProjects tool (Status: 200, Content-Type: application/json)
const GetProjectsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the board's projects, at the specified page of the results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **projectCategory** (Type: object):\n    - **description** (Type: string):\n        - Example: 'This is a project category'\n    - **id** (Type: string):\n        - Example: '10000'\n    - **name** (Type: string):\n        - Example: 'My Project Category'\n    - **self** (Type: string):\n        - Example: 'http://www.example.com/jira/rest/api/2/projectCategory/10000'\n  - **projectTypeKey** (Type: string):\n  - **self** (Type: string):\n  - **avatarUrls** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: string):\n  - **id** (Type: string):\n  - **key** (Type: string):\n  - **name** (Type: string):\n"

// NewGetProjectsMCPTool creates the MCP Tool instance for GetProjects
func NewGetProjectsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProjects",
		"Get all projects associated with the board - Returns all projects that are associated with the board, for the given board Id. A project is associated with a board only if the board filter explicitly filters issues by the project and guaranties that all issues will come for one of those projects e.g. board's filter with \"project in (PR-1, PR-1) OR reporter = admin\" jql Projects are returned only if user can browse all projects that are associated with the board. Note, if the user does not have permission to view the board, no projects will be returned at all. Returned projects are ordered by the name.",
		[]byte(GetProjectsInputSchema),
	)
}

// GetProjectsHandler is the handler function for the GetProjects tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/project", args, []string{"boardId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetProjects")
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
