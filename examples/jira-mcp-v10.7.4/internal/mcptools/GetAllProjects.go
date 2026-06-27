package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAllProjects tool
const GetAllProjectsInputSchema = "{\n  \"properties\": {\n    \"browseArchive\": {\n      \"description\": \"Whether to include only projects where current user can browse archive\",\n      \"type\": \"boolean\"\n    },\n    \"expand\": {\n      \"description\": \"Parameters to expand\",\n      \"type\": \"string\"\n    },\n    \"includeArchived\": {\n      \"description\": \"Whether to include archived projects in response, default: false\",\n      \"type\": \"boolean\"\n    },\n    \"recent\": {\n      \"description\": \"If this parameter is set then only projects recently accessed by the current user (if not logged in then based on HTTP session) will be returned (maximum count limited to the specified number but no more than 20)\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetAllProjects tool (Status: 200, Content-Type: application/json)
const GetAllProjectsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Project data\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'Example'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/project/EX'\n  - **archived** (Type: boolean):\n      - Example: 'false'\n  - **avatarUrls** (Type: object):\n      - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n  - **description** (Type: string):\n      - Example: 'Example'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **key** (Type: string):\n      - Example: 'EX'\n"

// NewGetAllProjectsMCPTool creates the MCP Tool instance for GetAllProjects
func NewGetAllProjectsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllProjects",
		"Get all visible projects - Returns all projects which are visible for the currently logged in user. If no user is logged in, it returns the list of projects that are visible when using anonymous access.",
		[]byte(GetAllProjectsInputSchema),
	)
}

// GetAllProjectsHandler is the handler function for the GetAllProjects tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/project", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAllProjects"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
