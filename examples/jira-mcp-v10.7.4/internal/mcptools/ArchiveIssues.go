package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the ArchiveIssues tool
const ArchiveIssuesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"List of issue keys\",\n      \"type\": \"string\"\n    },\n    \"notifyUsers\": {\n      \"description\": \"Send the email with notification that the issue was updated to users that watch it. Admin or project admin permissions are required to disable the notification.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the ArchiveIssues tool (Status: 200, Content-Type: text/plain)
const ArchiveIssuesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** text/plain\n\n> Returns a stream of issues archiving results.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewArchiveIssuesMCPTool creates the MCP Tool instance for ArchiveIssues
func NewArchiveIssuesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ArchiveIssues",
		"Archive list of issues - Archives a list of issues.",
		[]byte(ArchiveIssuesInputSchema),
	)
}

// ArchiveIssuesHandler is the handler function for the ArchiveIssues tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ArchiveIssuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "text/plain"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/issue/archive", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ArchiveIssues"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
