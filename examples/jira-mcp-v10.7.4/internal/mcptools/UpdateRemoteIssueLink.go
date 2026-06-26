package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateRemoteIssueLink tool
const UpdateRemoteIssueLinkInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Remote issue link create or update request\",\n      \"properties\": {\n        \"application\": {\n          \"properties\": {\n            \"name\": {\n              \"example\": \"My Acme Tracker\",\n              \"type\": \"string\"\n            },\n            \"type\": {\n              \"example\": \"com.acme.tracker\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"globalId\": {\n          \"example\": \"system=http://www.mycompany.com/support\\u0026id=1\",\n          \"type\": \"string\"\n        },\n        \"object\": {\n          \"properties\": {\n            \"icon\": {\n              \"example\": \"http://www.mycompany.com/support/resolved.png\",\n              \"properties\": {\n                \"link\": {\n                  \"example\": \"http://www.mycompany.com/support/resolved.png\",\n                  \"type\": \"string\"\n                },\n                \"title\": {\n                  \"example\": \"Support Ticket\",\n                  \"type\": \"string\"\n                },\n                \"url16x16\": {\n                  \"example\": \"http://www.mycompany.com/support/ticket.png\",\n                  \"type\": \"string\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"status\": {\n              \"properties\": {\n                \"icon\": {},\n                \"resolved\": {\n                  \"example\": true,\n                  \"type\": \"boolean\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"summary\": {\n              \"example\": \"Crazy customer support issue\",\n              \"type\": \"string\"\n            },\n            \"title\": {\n              \"example\": \"TSTSUP-111\",\n              \"type\": \"string\"\n            },\n            \"url\": {\n              \"example\": \"http://www.mycompany.com/support?id=1\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"relationship\": {\n          \"example\": \"causes\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    },\n    \"linkId\": {\n      \"description\": \"Id of the remote issue link\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\",\n    \"linkId\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateRemoteIssueLinkMCPTool creates the MCP Tool instance for UpdateRemoteIssueLink
func NewUpdateRemoteIssueLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateRemoteIssueLink",
		"Update remote issue link - Updates a remote issue link from a JSON representation. Any fields not provided are set to null.",
		[]byte(UpdateRemoteIssueLinkInputSchema),
	)
}

// UpdateRemoteIssueLinkHandler is the handler function for the UpdateRemoteIssueLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateRemoteIssueLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issue/{issueIdOrKey}/remotelink/{linkId}", args, []string{"issueIdOrKey", "linkId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateRemoteIssueLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
