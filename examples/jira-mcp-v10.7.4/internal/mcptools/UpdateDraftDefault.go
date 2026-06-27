package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateDraftDefault tool
const UpdateDraftDefaultInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The new default.\",\n      \"properties\": {\n        \"updateDraftIfNeeded\": {\n          \"type\": \"boolean\"\n        },\n        \"workflow\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"The id of the parent scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateDraftDefault tool (Status: 200, Content-Type: application/json)
const UpdateDraftDefaultResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned on success.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **lastModifiedUser** (Type: object):\n    - **expand** (Type: string):\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n    - **key** (Type: string):\n        - Example: 'JIRAUSER10100'\n    - **locale** (Type: string):\n        - Example: 'en_AU'\n    - **applicationRoles** (Type: object):\n        - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n      - **size** (Type: integer, int32):\n      - **callback** (Type: object):\n      - **maxResults** (Type: integer, int32):\n      - **[cyclic reference]**\n    - **avatarUrls** (Type: object):\n        - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n      - **Additional Properties**:\n        - **property value** (Type: string, uri):\n            - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n    - **deleted** (Type: boolean):\n        - Example: 'false'\n    - **name** (Type: string):\n        - Example: 'fred'\n    - **self** (Type: string, uri):\n        - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **groups** (Type: object):\n      - **callback** (Type: object):\n      - **maxResults** (Type: integer, int32):\n      - **[cyclic reference]**\n      - **size** (Type: integer, int32):\n    - **lastLoginTime** (Type: string):\n        - Example: '2023-08-30T16:37:01+1000'\n  - **updateDraftIfNeeded** (Type: boolean):\n      - Example: 'true'\n  - **draft** (Type: boolean):\n      - Example: 'false'\n  - **originalIssueTypeMappings** (Type: object):\n      - Example: '{\"IssueTypeId\":\"WorkflowName2\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: '{\"IssueTypeId\":\"WorkflowName2\"}'\n  - **id** (Type: integer, int64):\n      - Example: '10000'\n  - **issueTypeMappings** (Type: object):\n      - Example: '{\"IsueTypeId\":\"WorkflowName\",\"IsueTypeId2\":\"WorkflowName\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: '{\"IsueTypeId\":\"WorkflowName\",\"IsueTypeId2\":\"WorkflowName\"}'\n  - **originalDefaultWorkflow** (Type: string):\n      - Example: 'ParentsDefaultWorkflowName'\n  - **defaultWorkflow** (Type: string):\n      - Example: 'DefaultWorkflowName'\n  - **description** (Type: string):\n      - Example: 'This is a workflow scheme'\n  - **issueTypes** (Type: object):\n      - Example: '{\"IsueTypeId\":{\"description\":\"IssueTypeDescription\",\"name\":\"IssueTypeName\"}}'\n    - **Additional Properties**:\n      - **property value** (Type: object):\n        - **subtask** (Type: boolean):\n            - Example: 'false'\n        - **avatarId** (Type: integer, int64):\n            - Example: '10002'\n        - **description** (Type: string):\n            - Example: 'A problem which impairs or prevents the functions of the product.'\n        - **iconUrl** (Type: string):\n            - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n        - **id** (Type: string):\n            - Example: '1'\n        - **name** (Type: string):\n            - Example: 'Bug'\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n  - **name** (Type: string):\n      - Example: 'My Workflow Scheme'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/workflowscheme/10000'\n  - **lastModified** (Type: string):\n      - Example: 'Today 12:45'\n"

// NewUpdateDraftDefaultMCPTool creates the MCP Tool instance for UpdateDraftDefault
func NewUpdateDraftDefaultMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateDraftDefault",
		"Update default workflow for a draft scheme - Set the default workflow for the passed draft workflow scheme.",
		[]byte(UpdateDraftDefaultInputSchema),
	)
}

// UpdateDraftDefaultHandler is the handler function for the UpdateDraftDefault tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateDraftDefaultHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/workflowscheme/{id}/draft/default", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateDraftDefault"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
