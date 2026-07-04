package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetProjectPullRequestsList tool
const GetProjectPullRequestsListInputSchema = "{\n  \"properties\": {\n    \"project\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"project\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProjectPullRequestsList tool (Status: 200, Content-Type: application/json)
const GetProjectPullRequestsListResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetProjectPullRequestsListMCPTool creates the MCP Tool instance for GetProjectPullRequestsList
func NewGetProjectPullRequestsListMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProjectPullRequestsList",
		"List the pull requests of a project. - List the pull requests of a project. One of the following permissions is required: - 'Browse' rights on the specified project - 'Execute Analysis' rights on the specified project",
		[]byte(GetProjectPullRequestsListInputSchema),
	)
}

// GetProjectPullRequestsListHandler is the handler function for the GetProjectPullRequestsList tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProjectPullRequestsListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/project_pull_requests/list", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetProjectPullRequestsList")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
