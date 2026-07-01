package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the RankIssues tool
const RankIssuesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Bean which contains list of issues to rank and information where it should be ranked.\",\n      \"properties\": {\n        \"issues\": {\n          \"example\": [\n            \"PR-1\",\n            \"10001\",\n            \"PR-3\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"PR-1\\\",\\\"10001\\\",\\\"PR-3\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"rankAfterIssue\": {\n          \"example\": \"PR-4\",\n          \"type\": \"string\"\n        },\n        \"rankBeforeIssue\": {\n          \"example\": \"PR-4\",\n          \"type\": \"string\"\n        },\n        \"rankCustomFieldId\": {\n          \"example\": 10521,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the RankIssues tool (Status: 207, Content-Type: application/json)
const RankIssuesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 207\n\n**Content-Type:** application/json\n\n> Returns the list of issue with status of rank operation.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **entries** (Type: array):\n    - **Items** (Type: object):\n      - **status** (Type: integer, int32):\n          - Example: '200'\n      - **errors** (Type: array):\n          - Example: '[\"JIRA Agile cannot execute the rank operation at this time. Please try again later.\"]'\n        - **Items** (Type: string):\n            - Example: '[\"JIRA Agile cannot execute the rank operation at this time. Please try again later.\"]'\n      - **issueId** (Type: integer, int64):\n          - Example: '10000'\n      - **issueKey** (Type: string):\n          - Example: 'PR-1'\n"

// NewRankIssuesMCPTool creates the MCP Tool instance for RankIssues
func NewRankIssuesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RankIssues",
		"Rank issues before or after a given issue - Moves (ranks) issues before or after a given issue. At most 50 issues may be ranked at once. This operation may fail for some issues, although this will be rare. In that case the 207 status code is returned for the whole response and detailed information regarding each issue is available in the response body. If rankCustomFieldId is not defined, the default rank field will be used.",
		[]byte(RankIssuesInputSchema),
	)
}

// RankIssuesHandler is the handler function for the RankIssues tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RankIssuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/agile/1.0/issue/rank", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "RankIssues")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
