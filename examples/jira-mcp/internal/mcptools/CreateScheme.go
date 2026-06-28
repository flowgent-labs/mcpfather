package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateScheme tool
const CreateSchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The body contains a representation of the new scheme. Values not passed are assumed to be set to their defaults.\",\n      \"properties\": {\n        \"defaultWorkflow\": {\n          \"example\": \"DefaultWorkflowName\",\n          \"type\": \"string\"\n        },\n        \"description\": {\n          \"example\": \"This is a workflow scheme\",\n          \"type\": \"string\"\n        },\n        \"draft\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"id\": {\n          \"example\": 10000,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"issueTypeMappings\": {\n          \"additionalProperties\": {\n            \"example\": \"{\\\"IsueTypeId\\\":\\\"WorkflowName\\\",\\\"IsueTypeId2\\\":\\\"WorkflowName\\\"}\",\n            \"type\": \"string\"\n          },\n          \"example\": {\n            \"IsueTypeId\": \"WorkflowName\",\n            \"IsueTypeId2\": \"WorkflowName\"\n          },\n          \"type\": \"object\"\n        },\n        \"issueTypes\": {\n          \"additionalProperties\": {\n            \"properties\": {\n              \"avatarId\": {\n                \"example\": 10002,\n                \"format\": \"int64\",\n                \"type\": \"integer\"\n              },\n              \"description\": {\n                \"example\": \"A problem which impairs or prevents the functions of the product.\",\n                \"type\": \"string\"\n              },\n              \"iconUrl\": {\n                \"example\": \"http://www.example.com/jira/images/icons/issuetypes/bug.png\",\n                \"type\": \"string\"\n              },\n              \"id\": {\n                \"example\": \"1\",\n                \"type\": \"string\"\n              },\n              \"name\": {\n                \"example\": \"Bug\",\n                \"type\": \"string\"\n              },\n              \"self\": {\n                \"example\": \"http://www.example.com/jira/rest/api/2/issuetype/1\",\n                \"type\": \"string\"\n              },\n              \"subtask\": {\n                \"example\": false,\n                \"type\": \"boolean\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"example\": {\n            \"IsueTypeId\": {\n              \"description\": \"IssueTypeDescription\",\n              \"name\": \"IssueTypeName\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"lastModified\": {\n          \"example\": \"Today 12:45\",\n          \"type\": \"string\"\n        },\n        \"lastModifiedUser\": {\n          \"properties\": {\n            \"active\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"applicationRoles\": {\n              \"example\": [\n                \"jira-core\",\n                \"jira-admin\",\n                \"important\"\n              ],\n              \"properties\": {\n                \"callback\": {},\n                \"maxResults\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                },\n                \"pagingCallback\": {\n                  \"type\": \"object\"\n                },\n                \"size\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"avatarUrls\": {\n              \"additionalProperties\": {\n                \"example\": \"{\\\"48x48\\\":\\\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\\\",\\\"24x24\\\":\\\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\\\",\\\"16x16\\\":\\\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\\\",\\\"32x32\\\":\\\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\\\"}\",\n                \"format\": \"uri\",\n                \"type\": \"string\"\n              },\n              \"example\": {\n                \"16x16\": \"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\n                \"24x24\": \"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\n                \"32x32\": \"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\n                \"48x48\": \"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"\n              },\n              \"type\": \"object\"\n            },\n            \"deleted\": {\n              \"example\": false,\n              \"type\": \"boolean\"\n            },\n            \"displayName\": {\n              \"example\": \"Fred F. User\",\n              \"type\": \"string\"\n            },\n            \"emailAddress\": {\n              \"example\": \"fred@example.com\",\n              \"type\": \"string\"\n            },\n            \"expand\": {\n              \"type\": \"string\"\n            },\n            \"groups\": {\n              \"properties\": {\n                \"callback\": {\n                  \"type\": \"object\"\n                },\n                \"maxResults\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                },\n                \"pagingCallback\": {},\n                \"size\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"key\": {\n              \"example\": \"JIRAUSER10100\",\n              \"type\": \"string\"\n            },\n            \"lastLoginTime\": {\n              \"example\": \"2023-08-30T16:37:01+1000\",\n              \"type\": \"string\"\n            },\n            \"locale\": {\n              \"example\": \"en_AU\",\n              \"type\": \"string\"\n            },\n            \"name\": {\n              \"example\": \"fred\",\n              \"type\": \"string\"\n            },\n            \"self\": {\n              \"example\": \"http://www.example.com/jira/rest/api/2.0/user?username=fred\",\n              \"format\": \"uri\",\n              \"type\": \"string\"\n            },\n            \"timeZone\": {\n              \"example\": \"Australia/Sydney\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"name\": {\n          \"example\": \"My Workflow Scheme\",\n          \"type\": \"string\"\n        },\n        \"originalDefaultWorkflow\": {\n          \"example\": \"ParentsDefaultWorkflowName\",\n          \"type\": \"string\"\n        },\n        \"originalIssueTypeMappings\": {\n          \"additionalProperties\": {\n            \"example\": \"{\\\"IssueTypeId\\\":\\\"WorkflowName2\\\"}\",\n            \"type\": \"string\"\n          },\n          \"example\": {\n            \"IssueTypeId\": \"WorkflowName2\"\n          },\n          \"type\": \"object\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/workflowscheme/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"updateDraftIfNeeded\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// NewCreateSchemeMCPTool creates the MCP Tool instance for CreateScheme
func NewCreateSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateScheme",
		"Create a new workflow scheme - Create a new workflow scheme. The body contains a representation of the new scheme. Values not passed are assumed to be set to their defaults.",
		[]byte(CreateSchemeInputSchema),
	)
}

// CreateSchemeHandler is the handler function for the CreateScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/workflowscheme", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateScheme")
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
