package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the Notify tool
const NotifyInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Notification request\",\n      \"properties\": {\n        \"htmlBody\": {\n          \"example\": \"Lorem ipsum \\u003cstrong\\u003edolor\\u003c/strong\\u003e sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.\",\n          \"type\": \"string\"\n        },\n        \"restrict\": {\n          \"properties\": {\n            \"groups\": {\n              \"items\": {\n                \"properties\": {\n                  \"name\": {\n                    \"example\": \"jira-administrators\",\n                    \"type\": \"string\"\n                  },\n                  \"self\": {\n                    \"example\": \"http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators\",\n                    \"format\": \"uri\",\n                    \"type\": \"string\"\n                  }\n                },\n                \"type\": \"object\"\n              },\n              \"type\": \"array\"\n            },\n            \"permissions\": {\n              \"items\": {\n                \"description\": \"A map of permission keys to permission objects.\",\n                \"example\": {\n                  \"EDIT_ISSUE\": {\n                    \"description\": \"Ability to edit issues.\",\n                    \"havePermission\": true,\n                    \"id\": \"EDIT_ISSUE\",\n                    \"name\": \"Edit Issues\",\n                    \"type\": \"USER\"\n                  }\n                },\n                \"properties\": {\n                  \"description\": {\n                    \"type\": \"string\"\n                  },\n                  \"id\": {\n                    \"type\": \"string\"\n                  },\n                  \"key\": {\n                    \"type\": \"string\"\n                  },\n                  \"name\": {\n                    \"type\": \"string\"\n                  },\n                  \"type\": {\n                    \"enum\": [\n                      \"GLOBAL\",\n                      \"PROJECT\"\n                    ],\n                    \"type\": \"string\"\n                  }\n                },\n                \"type\": \"object\"\n              },\n              \"type\": \"array\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"subject\": {\n          \"example\": \"Duis eu justo eget augue iaculis fermentum.\",\n          \"type\": \"string\"\n        },\n        \"textBody\": {\n          \"example\": \"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque eget venenatis elit. Duis eu justo eget augue iaculis fermentum. Sed semper quam laoreet nisi egestas at posuere augue semper.\",\n          \"type\": \"string\"\n        },\n        \"to\": {\n          \"properties\": {\n            \"assignee\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"groups\": {\n              \"items\": {},\n              \"type\": \"array\"\n            },\n            \"reporter\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"users\": {\n              \"items\": {\n                \"properties\": {\n                  \"active\": {\n                    \"example\": true,\n                    \"type\": \"boolean\"\n                  },\n                  \"avatarUrls\": {\n                    \"additionalProperties\": {\n                      \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n                      \"type\": \"string\"\n                    },\n                    \"example\": \"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\",\n                    \"type\": \"object\"\n                  },\n                  \"displayName\": {\n                    \"example\": \"Fred F. User\",\n                    \"type\": \"string\"\n                  },\n                  \"emailAddress\": {\n                    \"example\": \"fred@example.com\",\n                    \"type\": \"string\"\n                  },\n                  \"key\": {\n                    \"example\": \"fred\",\n                    \"type\": \"string\"\n                  },\n                  \"name\": {\n                    \"example\": \"Fred\",\n                    \"type\": \"string\"\n                  },\n                  \"self\": {\n                    \"example\": \"http://www.example.com/jira/rest/api/2/user?username=fred\",\n                    \"type\": \"string\"\n                  },\n                  \"timeZone\": {\n                    \"example\": \"Australia/Sydney\",\n                    \"type\": \"string\"\n                  }\n                },\n                \"type\": \"object\"\n              },\n              \"type\": \"array\"\n            },\n            \"voters\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            },\n            \"watchers\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            }\n          },\n          \"type\": \"object\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewNotifyMCPTool creates the MCP Tool instance for Notify
func NewNotifyMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Notify",
		"Send notification to recipients - Sends a notification (email) to the list or recipients defined in the request.",
		[]byte(NotifyInputSchema),
	)
}

// NotifyHandler is the handler function for the Notify tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func NotifyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issue/{issueIdOrKey}/notify", args, []string{"issueIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Notify"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
