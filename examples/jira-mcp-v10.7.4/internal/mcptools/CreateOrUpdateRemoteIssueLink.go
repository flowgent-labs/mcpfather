package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the CreateOrUpdateRemoteIssueLink tool
const CreateOrUpdateRemoteIssueLinkInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Remote issue link create or update request\",\n      \"properties\": {\n        \"application\": {\n          \"properties\": {\n            \"name\": {\n              \"example\": \"My Acme Tracker\",\n              \"type\": \"string\"\n            },\n            \"type\": {\n              \"example\": \"com.acme.tracker\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"globalId\": {\n          \"example\": \"system=http://www.mycompany.com/support\\u0026id=1\",\n          \"type\": \"string\"\n        },\n        \"object\": {\n          \"properties\": {\n            \"icon\": {\n              \"example\": \"http://www.mycompany.com/support/resolved.png\",\n              \"properties\": {\n                \"link\": {\n                  \"example\": \"http://www.mycompany.com/support/resolved.png\",\n                  \"type\": \"string\"\n                },\n                \"title\": {\n                  \"example\": \"Support Ticket\",\n                  \"type\": \"string\"\n                },\n                \"url16x16\": {\n                  \"example\": \"http://www.mycompany.com/support/ticket.png\",\n                  \"type\": \"string\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"status\": {\n              \"properties\": {\n                \"icon\": {},\n                \"resolved\": {\n                  \"example\": true,\n                  \"type\": \"boolean\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"summary\": {\n              \"example\": \"Crazy customer support issue\",\n              \"type\": \"string\"\n            },\n            \"title\": {\n              \"example\": \"TSTSUP-111\",\n              \"type\": \"string\"\n            },\n            \"url\": {\n              \"example\": \"http://www.mycompany.com/support?id=1\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"relationship\": {\n          \"example\": \"causes\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateOrUpdateRemoteIssueLink tool (Status: 200, Content-Type: application/json)
const CreateOrUpdateRemoteIssueLinkResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a link to the created/updated remote issue link.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewCreateOrUpdateRemoteIssueLinkMCPTool creates the MCP Tool instance for CreateOrUpdateRemoteIssueLink
func NewCreateOrUpdateRemoteIssueLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateOrUpdateRemoteIssueLink",
		"Create or update remote issue link - Creates or updates a remote issue link from a JSON representation. If a globalId is provided and a remote issue link exists with that globalId, the remote issue link is updated. Otherwise, the remote issue link is created.",
		[]byte(CreateOrUpdateRemoteIssueLinkInputSchema),
	)
}

// CreateOrUpdateRemoteIssueLinkHandler is the handler function for the CreateOrUpdateRemoteIssueLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateOrUpdateRemoteIssueLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issue/{issueIdOrKey}/remotelink", args, []string{"issueIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateOrUpdateRemoteIssueLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
