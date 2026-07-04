package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetAlmIntegrationsListBitbucketserverProjects tool
const GetAlmIntegrationsListBitbucketserverProjectsInputSchema = "{\n  \"properties\": {\n    \"almSetting\": {\n      \"description\": \"DevOps Platform setting key\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    },\n    \"pageSize\": {\n      \"description\": \"Number of items to return.\",\n      \"maximum\": 100,\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"Start number for the page (inclusive). If not passed, the first page is assumed.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"almSetting\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAlmIntegrationsListBitbucketserverProjects tool (Status: 200, Content-Type: application/json)
const GetAlmIntegrationsListBitbucketserverProjectsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetAlmIntegrationsListBitbucketserverProjectsMCPTool creates the MCP Tool instance for GetAlmIntegrationsListBitbucketserverProjects
func NewGetAlmIntegrationsListBitbucketserverProjectsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAlmIntegrationsListBitbucketserverProjects",
		"List the Bitbucket Server projects. - List the Bitbucket Server projects Requires the 'Create Projects' permission",
		[]byte(GetAlmIntegrationsListBitbucketserverProjectsInputSchema),
	)
}

// GetAlmIntegrationsListBitbucketserverProjectsHandler is the handler function for the GetAlmIntegrationsListBitbucketserverProjects tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAlmIntegrationsListBitbucketserverProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/alm_integrations/list_bitbucketserver_projects", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAlmIntegrationsListBitbucketserverProjects")
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
