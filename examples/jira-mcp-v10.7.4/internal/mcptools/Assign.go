package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the Assign tool
const AssignInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"UserBean containing the username\",\n      \"properties\": {\n        \"active\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"applicationRoles\": {\n          \"example\": [\n            \"jira-core\",\n            \"jira-admin\",\n            \"important\"\n          ],\n          \"properties\": {\n            \"callback\": {\n              \"type\": \"object\"\n            },\n            \"maxResults\": {\n              \"format\": \"int32\",\n              \"type\": \"integer\"\n            },\n            \"pagingCallback\": {},\n            \"size\": {\n              \"format\": \"int32\",\n              \"type\": \"integer\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"avatarUrls\": {\n          \"additionalProperties\": {\n            \"example\": \"{\\\"48x48\\\":\\\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\\\",\\\"24x24\\\":\\\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\\\",\\\"16x16\\\":\\\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\\\",\\\"32x32\\\":\\\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\\\"}\",\n            \"format\": \"uri\",\n            \"type\": \"string\"\n          },\n          \"example\": {\n            \"16x16\": \"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\n            \"24x24\": \"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\n            \"32x32\": \"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\n            \"48x48\": \"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"\n          },\n          \"type\": \"object\"\n        },\n        \"deleted\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"displayName\": {\n          \"example\": \"Fred F. User\",\n          \"type\": \"string\"\n        },\n        \"emailAddress\": {\n          \"example\": \"fred@example.com\",\n          \"type\": \"string\"\n        },\n        \"expand\": {\n          \"type\": \"string\"\n        },\n        \"groups\": {\n          \"properties\": {\n            \"callback\": {},\n            \"maxResults\": {\n              \"format\": \"int32\",\n              \"type\": \"integer\"\n            },\n            \"pagingCallback\": {\n              \"type\": \"object\"\n            },\n            \"size\": {\n              \"format\": \"int32\",\n              \"type\": \"integer\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"key\": {\n          \"example\": \"JIRAUSER10100\",\n          \"type\": \"string\"\n        },\n        \"lastLoginTime\": {\n          \"example\": \"2023-08-30T16:37:01+1000\",\n          \"type\": \"string\"\n        },\n        \"locale\": {\n          \"example\": \"en_AU\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"fred\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2.0/user?username=fred\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"timeZone\": {\n          \"example\": \"Australia/Sydney\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewAssignMCPTool creates the MCP Tool instance for Assign
func NewAssignMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Assign",
		"Assign an issue to a user - Assign an issue to a user.",
		[]byte(AssignInputSchema),
	)
}

// AssignHandler is the handler function for the Assign tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AssignHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issue/{issueIdOrKey}/assignee", args, []string{"issueIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Assign"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
