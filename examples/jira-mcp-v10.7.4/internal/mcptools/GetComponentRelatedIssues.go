package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetComponentRelatedIssues tool
const GetComponentRelatedIssuesInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"a String containing the component id\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetComponentRelatedIssues tool (Status: 200, Content-Type: application/json)
const GetComponentRelatedIssuesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns counts of issues related to this component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **issueCount** (Type: integer, int64):\n      - Example: '23'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/component/10000'\n"

// NewGetComponentRelatedIssuesMCPTool creates the MCP Tool instance for GetComponentRelatedIssues
func NewGetComponentRelatedIssuesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetComponentRelatedIssues",
		"Get component related issues - Returns counts of issues related to this component.",
		[]byte(GetComponentRelatedIssuesInputSchema),
	)
}

// GetComponentRelatedIssuesHandler is the handler function for the GetComponentRelatedIssues tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetComponentRelatedIssuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/component/{id}/relatedIssueCounts", args, []string{"id"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetComponentRelatedIssues"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
