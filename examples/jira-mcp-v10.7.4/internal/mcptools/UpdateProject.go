package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateProject tool
const UpdateProjectInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Project update data\",\n      \"properties\": {\n        \"assigneeType\": {\n          \"enum\": [\n            \"PROJECT_LEAD\",\n            \"UNASSIGNED\"\n          ],\n          \"example\": \"PROJECT_LEAD\",\n          \"type\": \"string\"\n        },\n        \"avatarId\": {\n          \"example\": 10200,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"categoryId\": {\n          \"example\": 10120,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"description\": {\n          \"example\": \"Example Project description\",\n          \"type\": \"string\"\n        },\n        \"issueSecurityScheme\": {\n          \"example\": 10001,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"key\": {\n          \"example\": \"EX\",\n          \"type\": \"string\"\n        },\n        \"lead\": {\n          \"example\": \"Charlie\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"Example\",\n          \"type\": \"string\"\n        },\n        \"notificationScheme\": {\n          \"example\": 10021,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"permissionScheme\": {\n          \"example\": 10011,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"projectTemplateKey\": {\n          \"example\": \"com.atlassian.jira-core-project-templates:jira-core-project-management\",\n          \"type\": \"string\"\n        },\n        \"projectTypeKey\": {\n          \"example\": \"business\",\n          \"type\": \"string\"\n        },\n        \"url\": {\n          \"example\": \"http://atlassian.com\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"expand\": {\n      \"description\": \"Parameters to expand\",\n      \"type\": \"string\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"Project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateProject tool (Status: 200, Content-Type: application/json)
const UpdateProjectResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Updated project data\n\n## Response Structure\n\n- Structure (Type: object):\n  - **avatarUrls** (Type: object):\n      - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n  - **description** (Type: string):\n      - Example: 'Example'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **key** (Type: string):\n      - Example: 'EX'\n  - **name** (Type: string):\n      - Example: 'Example'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/project/EX'\n  - **archived** (Type: boolean):\n      - Example: 'false'\n"

// NewUpdateProjectMCPTool creates the MCP Tool instance for UpdateProject
func NewUpdateProjectMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateProject",
		"Update a project - Updates a project. Only non null values sent in JSON will be updated in the project. Values available for the assigneeType field are: \"PROJECT_LEAD\" and \"UNASSIGNED\".",
		[]byte(UpdateProjectInputSchema),
	)
}

// UpdateProjectHandler is the handler function for the UpdateProject tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/project/{projectIdOrKey}", args, []string{"projectIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateProject"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
