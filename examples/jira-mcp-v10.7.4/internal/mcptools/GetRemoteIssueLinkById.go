package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetRemoteIssueLinkById tool
const GetRemoteIssueLinkByIdInputSchema = "{\n  \"properties\": {\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    },\n    \"linkId\": {\n      \"description\": \"Id of the remote issue link\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\",\n    \"linkId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRemoteIssueLinkById tool (Status: 200, Content-Type: application/json)
const GetRemoteIssueLinkByIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a response containing a remote issue link.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewGetRemoteIssueLinkByIdMCPTool creates the MCP Tool instance for GetRemoteIssueLinkById
func NewGetRemoteIssueLinkByIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRemoteIssueLinkById",
		"Get a remote issue link by its id - Get a remote issue link by its id.",
		[]byte(GetRemoteIssueLinkByIdInputSchema),
	)
}

// GetRemoteIssueLinkByIdHandler is the handler function for the GetRemoteIssueLinkById tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRemoteIssueLinkByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issue/{issueIdOrKey}/remotelink/{linkId}", args, []string{"issueIdOrKey", "linkId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetRemoteIssueLinkById"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
