package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetRemoteVersionLinksByVersionId tool
const GetRemoteVersionLinksByVersionIdInputSchema = "{\n  \"properties\": {\n    \"versionId\": {\n      \"description\": \"ID of the version.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"versionId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRemoteVersionLinksByVersionId tool (Status: 200, Content-Type: application/json)
const GetRemoteVersionLinksByVersionIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the version exists and the currently authenticated user has permission to view it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **links** (Type: array):\n    - **Items** (Type: object):\n      - **name** (Type: string):\n          - Example: 'Issue 10000'\n      - **self** (Type: string, uri):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/10000'\n      - **link** (Type: string):\n          - Example: '{\"rel\":\"issue\",\"url\":\"http://www.example.com/jira/rest/api/2/issue/10000\"}'\n"

// NewGetRemoteVersionLinksByVersionIdMCPTool creates the MCP Tool instance for GetRemoteVersionLinksByVersionId
func NewGetRemoteVersionLinksByVersionIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRemoteVersionLinksByVersionId",
		"Get remote version links by version ID - Returns the remote version links associated with the given version ID.",
		[]byte(GetRemoteVersionLinksByVersionIdInputSchema),
	)
}

// GetRemoteVersionLinksByVersionIdHandler is the handler function for the GetRemoteVersionLinksByVersionId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRemoteVersionLinksByVersionIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/version/{versionId}/remotelink", args, []string{"versionId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetRemoteVersionLinksByVersionId"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
