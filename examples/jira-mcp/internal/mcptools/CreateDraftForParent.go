package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateDraftForParent tool
const CreateDraftForParentInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The id of the parent scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateDraftForParent tool (Status: 201, Content-Type: application/json)
const CreateDraftForParentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> The newly created scheme.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **originalDefaultWorkflow** (Type: string):\n      - Example: 'ParentsDefaultWorkflowName'\n  - **originalIssueTypeMappings** (Type: object):\n      - Example: '{\"IssueTypeId\":\"WorkflowName2\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: '{\"IssueTypeId\":\"WorkflowName2\"}'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/workflowscheme/10000'\n  - **updateDraftIfNeeded** (Type: boolean):\n      - Example: 'true'\n  - **defaultWorkflow** (Type: string):\n      - Example: 'DefaultWorkflowName'\n  - **description** (Type: string):\n      - Example: 'This is a workflow scheme'\n  - **draft** (Type: boolean):\n      - Example: 'false'\n  - **id** (Type: integer, int64):\n      - Example: '10000'\n  - **issueTypeMappings** (Type: object):\n      - Example: '{\"IsueTypeId\":\"WorkflowName\",\"IsueTypeId2\":\"WorkflowName\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: '{\"IsueTypeId\":\"WorkflowName\",\"IsueTypeId2\":\"WorkflowName\"}'\n  - **lastModifiedUser** (Type: object):\n    - **deleted** (Type: boolean):\n        - Example: 'false'\n    - **groups** (Type: object):\n      - **callback** (Type: object):\n      - **maxResults** (Type: integer, int32):\n      - **[cyclic reference]**\n      - **size** (Type: integer, int32):\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **applicationRoles** (Type: object):\n        - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n      - **maxResults** (Type: integer, int32):\n      - **pagingCallback** (Type: object):\n      - **size** (Type: integer, int32):\n      - **[cyclic reference]**\n    - **avatarUrls** (Type: object):\n        - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n      - **Additional Properties**:\n        - **property value** (Type: string, uri):\n            - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **expand** (Type: string):\n    - **locale** (Type: string):\n        - Example: 'en_AU'\n    - **key** (Type: string):\n        - Example: 'JIRAUSER10100'\n    - **name** (Type: string):\n        - Example: 'fred'\n    - **self** (Type: string, uri):\n        - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n    - **lastLoginTime** (Type: string):\n        - Example: '2023-08-30T16:37:01+1000'\n  - **name** (Type: string):\n      - Example: 'My Workflow Scheme'\n  - **issueTypes** (Type: object):\n      - Example: '{\"IsueTypeId\":{\"description\":\"IssueTypeDescription\",\"name\":\"IssueTypeName\"}}'\n    - **Additional Properties**:\n      - **property value** (Type: object):\n        - **id** (Type: string):\n            - Example: '1'\n        - **name** (Type: string):\n            - Example: 'Bug'\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n        - **subtask** (Type: boolean):\n            - Example: 'false'\n        - **avatarId** (Type: integer, int64):\n            - Example: '10002'\n        - **description** (Type: string):\n            - Example: 'A problem which impairs or prevents the functions of the product.'\n        - **iconUrl** (Type: string):\n            - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n  - **lastModified** (Type: string):\n      - Example: 'Today 12:45'\n"

// NewCreateDraftForParentMCPTool creates the MCP Tool instance for CreateDraftForParent
func NewCreateDraftForParentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateDraftForParent",
		"Create a draft for a workflow scheme - Create a draft for the passed scheme. The draft will be a copy of the state of the parent.",
		[]byte(CreateDraftForParentInputSchema),
	)
}

// CreateDraftForParentHandler is the handler function for the CreateDraftForParent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateDraftForParentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/workflowscheme/{id}/createdraft", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateDraftForParent")
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
