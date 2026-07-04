package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetIssuesAuthors tool
const GetIssuesAuthorsInputSchema = "{\n  \"properties\": {\n    \"project\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    },\n    \"ps\": {\n      \"description\": \"Page size. Must be greater than 0 and less or equal than 100\",\n      \"maximum\": 100,\n      \"type\": \"integer\"\n    },\n    \"q\": {\n      \"description\": \"Limit search to authors that contain the supplied string.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetIssuesAuthors tool (Status: 200, Content-Type: application/json)
const GetIssuesAuthorsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetIssuesAuthorsMCPTool creates the MCP Tool instance for GetIssuesAuthors
func NewGetIssuesAuthorsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssuesAuthors",
		"Search SCM accounts which match a given query. - Search SCM accounts which match a given query. Requires authentication. When issue indexing is in progress returns 503 service unavailable HTTP code.",
		[]byte(GetIssuesAuthorsInputSchema),
	)
}

// GetIssuesAuthorsHandler is the handler function for the GetIssuesAuthors tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssuesAuthorsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/issues/authors", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetIssuesAuthors")
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
