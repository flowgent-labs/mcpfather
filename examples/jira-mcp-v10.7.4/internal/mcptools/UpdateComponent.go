package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateComponent tool
const UpdateComponentInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"the request containing value of the component's property. The value has to be a valid, non-empty JSON conforming to http://tools.ietf.org/html/rfc4627. The maximum length of the property value is 32768 bytes.\",\n      \"properties\": {\n        \"archived\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"assigneeType\": {\n          \"enum\": [\n            \"PROJECT_DEFAULT\",\n            \"COMPONENT_LEAD\",\n            \"PROJECT_LEAD\",\n            \"UNASSIGNED\"\n          ],\n          \"example\": \"PROJECT_LEAD\",\n          \"type\": \"string\"\n        },\n        \"deleted\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"description\": {\n          \"example\": \"This is a Jira component\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"lead\": {\n          \"properties\": {\n            \"active\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"applicationRoles\": {\n              \"example\": [\n                \"jira-core\",\n                \"jira-admin\",\n                \"important\"\n              ],\n              \"properties\": {\n                \"callback\": {\n                  \"type\": \"object\"\n                },\n                \"maxResults\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                },\n                \"pagingCallback\": {},\n                \"size\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"avatarUrls\": {\n              \"additionalProperties\": {\n                \"example\": \"{\\\"48x48\\\":\\\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\\\",\\\"24x24\\\":\\\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\\\",\\\"16x16\\\":\\\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\\\",\\\"32x32\\\":\\\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\\\"}\",\n                \"format\": \"uri\",\n                \"type\": \"string\"\n              },\n              \"example\": {\n                \"16x16\": \"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\n                \"24x24\": \"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\n                \"32x32\": \"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\n                \"48x48\": \"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"\n              },\n              \"type\": \"object\"\n            },\n            \"deleted\": {\n              \"example\": false,\n              \"type\": \"boolean\"\n            },\n            \"displayName\": {\n              \"example\": \"Fred F. User\",\n              \"type\": \"string\"\n            },\n            \"emailAddress\": {\n              \"example\": \"fred@example.com\",\n              \"type\": \"string\"\n            },\n            \"expand\": {\n              \"type\": \"string\"\n            },\n            \"groups\": {\n              \"properties\": {\n                \"callback\": {\n                  \"type\": \"object\"\n                },\n                \"maxResults\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                },\n                \"pagingCallback\": {},\n                \"size\": {\n                  \"format\": \"int32\",\n                  \"type\": \"integer\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"key\": {\n              \"example\": \"JIRAUSER10100\",\n              \"type\": \"string\"\n            },\n            \"lastLoginTime\": {\n              \"example\": \"2023-08-30T16:37:01+1000\",\n              \"type\": \"string\"\n            },\n            \"locale\": {\n              \"example\": \"en_AU\",\n              \"type\": \"string\"\n            },\n            \"name\": {\n              \"example\": \"fred\",\n              \"type\": \"string\"\n            },\n            \"self\": {\n              \"example\": \"http://www.example.com/jira/rest/api/2.0/user?username=fred\",\n              \"format\": \"uri\",\n              \"type\": \"string\"\n            },\n            \"timeZone\": {\n              \"example\": \"Australia/Sydney\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"leadUserName\": {\n          \"example\": \"fred\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"Component 1\",\n          \"type\": \"string\"\n        },\n        \"project\": {\n          \"example\": \"HSP\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/component/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"The component to delete.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateComponent tool (Status: 200, Content-Type: application/json)
const UpdateComponentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the component is successfully updated.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **leadUserName** (Type: string):\n      - Example: 'fred'\n  - **deleted** (Type: boolean):\n      - Example: 'false'\n  - **description** (Type: string):\n      - Example: 'This is a Jira component'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **name** (Type: string):\n      - Example: 'Component 1'\n  - **lead** (Type: object):\n    - **displayName** (Type: string):\n        - Example: 'Fred F. User'\n    - **key** (Type: string):\n        - Example: 'JIRAUSER10100'\n    - **locale** (Type: string):\n        - Example: 'en_AU'\n    - **applicationRoles** (Type: object):\n        - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n      - **callback** (Type: object):\n      - **maxResults** (Type: integer, int32):\n      - **[cyclic reference]**\n      - **size** (Type: integer, int32):\n    - **avatarUrls** (Type: object):\n        - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n      - **Additional Properties**:\n        - **property value** (Type: string, uri):\n            - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n    - **deleted** (Type: boolean):\n        - Example: 'false'\n    - **name** (Type: string):\n        - Example: 'fred'\n    - **self** (Type: string, uri):\n        - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n    - **active** (Type: boolean):\n        - Example: 'true'\n    - **emailAddress** (Type: string):\n        - Example: 'fred@example.com'\n    - **groups** (Type: object):\n      - **size** (Type: integer, int32):\n      - **callback** (Type: object):\n      - **maxResults** (Type: integer, int32):\n      - **[cyclic reference]**\n    - **lastLoginTime** (Type: string):\n        - Example: '2023-08-30T16:37:01+1000'\n    - **expand** (Type: string):\n    - **timeZone** (Type: string):\n        - Example: 'Australia/Sydney'\n  - **project** (Type: string):\n      - Example: 'HSP'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/component/10000'\n  - **archived** (Type: boolean):\n      - Example: 'false'\n  - **assigneeType** (Type: string):\n      - Example: 'PROJECT_LEAD'\n      - Enum: ['PROJECT_DEFAULT', 'COMPONENT_LEAD', 'PROJECT_LEAD', 'UNASSIGNED']\n"

// NewUpdateComponentMCPTool creates the MCP Tool instance for UpdateComponent
func NewUpdateComponentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateComponent",
		"Update a component - Modify a component via PUT. Any fields present in the PUT will override existing values. As a convenience, if a field is not present, it is silently ignored.",
		[]byte(UpdateComponentInputSchema),
	)
}

// UpdateComponentHandler is the handler function for the UpdateComponent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateComponentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/component/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateComponent"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
