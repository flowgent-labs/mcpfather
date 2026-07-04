package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostUsersUpdate tool
const PostUsersUpdateInputSchema = "{\n  \"properties\": {\n    \"email\": {\n      \"description\": \"User email\",\n      \"maxLength\": 100,\n      \"type\": \"string\"\n    },\n    \"login\": {\n      \"description\": \"User login\",\n      \"maxLength\": 255,\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"User name\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    },\n    \"scmAccount\": {\n      \"description\": \"SCM accounts. To set several values, the parameter must be called once for each value.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"login\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PostUsersUpdate tool (Status: 200, Content-Type: application/json)
const PostUsersUpdateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewPostUsersUpdateMCPTool creates the MCP Tool instance for PostUsersUpdate
func NewPostUsersUpdateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostUsersUpdate",
		"Update a user. - Update a user. Requires Administer System permission",
		[]byte(PostUsersUpdateInputSchema),
	)
}

// PostUsersUpdateHandler is the handler function for the PostUsersUpdate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostUsersUpdateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/users/update", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostUsersUpdate")
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
