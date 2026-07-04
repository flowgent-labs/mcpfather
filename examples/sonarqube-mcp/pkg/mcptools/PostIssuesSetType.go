package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostIssuesSetType tool
const PostIssuesSetTypeInputSchema = "{\n  \"properties\": {\n    \"issue\": {\n      \"description\": \"Issue key\",\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"description\": \"New type\",\n      \"enum\": [\n        \"CODE_SMELL\",\n        \"BUG\",\n        \"VULNERABILITY\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issue\",\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PostIssuesSetType tool (Status: 200, Content-Type: application/json)
const PostIssuesSetTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewPostIssuesSetTypeMCPTool creates the MCP Tool instance for PostIssuesSetType
func NewPostIssuesSetTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostIssuesSetType",
		"Change type of issue, for instance from 'code smell' to 'bug'. - Change type of issue, for instance from 'code smell' to 'bug'. Requires the following permissions: - 'Authentication' - 'Browse' rights on project of the specified issue - 'Administer Issues' rights on project of the specified issue",
		[]byte(PostIssuesSetTypeInputSchema),
	)
}

// PostIssuesSetTypeHandler is the handler function for the PostIssuesSetType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostIssuesSetTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/issues/set_type", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostIssuesSetType")
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
