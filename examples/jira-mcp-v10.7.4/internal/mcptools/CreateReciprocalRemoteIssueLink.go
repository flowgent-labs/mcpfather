package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the CreateReciprocalRemoteIssueLink tool
const CreateReciprocalRemoteIssueLinkInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Remote reciprocal issue link create request\",\n      \"properties\": {\n        \"source\": {\n          \"properties\": {\n            \"application\": {\n              \"properties\": {\n                \"name\": {\n                  \"example\": \"My Acme Tracker\",\n                  \"type\": \"string\"\n                },\n                \"type\": {\n                  \"example\": \"com.acme.tracker\",\n                  \"type\": \"string\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"globalId\": {\n              \"example\": \"system=http://www.mycompany.com/support\\u0026id=1\",\n              \"type\": \"string\"\n            },\n            \"object\": {\n              \"properties\": {\n                \"icon\": {\n                  \"example\": \"http://www.mycompany.com/support/resolved.png\",\n                  \"properties\": {\n                    \"link\": {\n                      \"example\": \"http://www.mycompany.com/support/resolved.png\",\n                      \"type\": \"string\"\n                    },\n                    \"title\": {\n                      \"example\": \"Support Ticket\",\n                      \"type\": \"string\"\n                    },\n                    \"url16x16\": {\n                      \"example\": \"http://www.mycompany.com/support/ticket.png\",\n                      \"type\": \"string\"\n                    }\n                  },\n                  \"type\": \"object\"\n                },\n                \"status\": {\n                  \"properties\": {\n                    \"icon\": {},\n                    \"resolved\": {\n                      \"example\": true,\n                      \"type\": \"boolean\"\n                    }\n                  },\n                  \"type\": \"object\"\n                },\n                \"summary\": {\n                  \"example\": \"Crazy customer support issue\",\n                  \"type\": \"string\"\n                },\n                \"title\": {\n                  \"example\": \"TSTSUP-111\",\n                  \"type\": \"string\"\n                },\n                \"url\": {\n                  \"example\": \"http://www.mycompany.com/support?id=1\",\n                  \"type\": \"string\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"relationship\": {\n              \"example\": \"causes\",\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"target\": {}\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the CreateReciprocalRemoteIssueLink tool (Status: 200, Content-Type: application/json)
const CreateReciprocalRemoteIssueLinkResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a pair of links to created remote issue links\n\n## Response Structure\n\n- Structure (Type: object):\n  - **source** (Type: object):\n  - **[cyclic reference]**\n"

// NewCreateReciprocalRemoteIssueLinkMCPTool creates the MCP Tool instance for CreateReciprocalRemoteIssueLink
func NewCreateReciprocalRemoteIssueLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateReciprocalRemoteIssueLink",
		"Create reciprocal remote issue link - Create reciprocal remote issue link from a JSON representation. Jira will create two issue links, source -> target and target -> source.",
		[]byte(CreateReciprocalRemoteIssueLinkInputSchema),
	)
}

// CreateReciprocalRemoteIssueLinkHandler is the handler function for the CreateReciprocalRemoteIssueLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateReciprocalRemoteIssueLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issue/remotelink/reciprocal", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateReciprocalRemoteIssueLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
