package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateProject tool
const CreateProjectInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Project data\",\n      \"properties\": {\n        \"assigneeType\": {\n          \"enum\": [\n            \"PROJECT_LEAD\",\n            \"UNASSIGNED\"\n          ],\n          \"example\": \"PROJECT_LEAD\",\n          \"type\": \"string\"\n        },\n        \"avatarId\": {\n          \"example\": 10200,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"categoryId\": {\n          \"example\": 10120,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"description\": {\n          \"example\": \"Example Project description\",\n          \"type\": \"string\"\n        },\n        \"issueSecurityScheme\": {\n          \"example\": 10001,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"key\": {\n          \"example\": \"EX\",\n          \"type\": \"string\"\n        },\n        \"lead\": {\n          \"example\": \"Charlie\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"Example\",\n          \"type\": \"string\"\n        },\n        \"notificationScheme\": {\n          \"example\": 10021,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"permissionScheme\": {\n          \"example\": 10011,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"projectTemplateKey\": {\n          \"example\": \"com.atlassian.jira-core-project-templates:jira-core-project-management\",\n          \"type\": \"string\"\n        },\n        \"projectTypeKey\": {\n          \"example\": \"business\",\n          \"type\": \"string\"\n        },\n        \"url\": {\n          \"example\": \"http://atlassian.com\",\n          \"type\": \"string\"\n        },\n        \"workflowSchemeId\": {\n          \"example\": 10031,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateProject tool (Status: 201, Content-Type: application/json)
const CreateProjectResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Created project data\n\n## Response Structure\n\n- Structure (Type: object):\n  - **key** (Type: string):\n      - Example: 'EX'\n  - **self** (Type: string, uri):\n      - Example: 'http://example/jira/rest/api/2/project/10042'\n  - **id** (Type: integer, int64):\n      - Example: '10010'\n"

// NewCreateProjectMCPTool creates the MCP Tool instance for CreateProject
func NewCreateProjectMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateProject",
		"Create a new project - Creates a new project",
		[]byte(CreateProjectInputSchema),
	)
}

// CreateProjectHandler is the handler function for the CreateProject tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/project", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateProject")
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
