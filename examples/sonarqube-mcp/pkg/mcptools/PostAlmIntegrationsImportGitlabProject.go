package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostAlmIntegrationsImportGitlabProject tool
const PostAlmIntegrationsImportGitlabProjectInputSchema = "{\n  \"properties\": {\n    \"almSetting\": {\n      \"description\": \"DevOps Platform configuration key. This parameter is optional if you have only one GitLab integration.\",\n      \"type\": \"string\"\n    },\n    \"gitlabProjectId\": {\n      \"description\": \"GitLab project ID\",\n      \"type\": \"string\"\n    },\n    \"newCodeDefinitionType\": {\n      \"description\": \"Project New Code Definition Type New code definitions of the following types are allowed: - PREVIOUS_VERSION - NUMBER_OF_DAYS - REFERENCE_BRANCH - will default to the main branch.\",\n      \"type\": \"string\"\n    },\n    \"newCodeDefinitionValue\": {\n      \"description\": \"Project New Code Definition Value For each new code definition type, a different value is expected: - no value, when the new code definition type is PREVIOUS_VERSION and REFERENCE_BRANCH - a number between 1 and 90, when the new code definition type is NUMBER_OF_DAYS\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"gitlabProjectId\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostAlmIntegrationsImportGitlabProjectMCPTool creates the MCP Tool instance for PostAlmIntegrationsImportGitlabProject
func NewPostAlmIntegrationsImportGitlabProjectMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostAlmIntegrationsImportGitlabProject",
		"Import a GitLab project to SonarQube, creating a new project and configuring MR decoration. - Import a GitLab project to SonarQube, creating a new project and configuring MR decoration Requires the 'Create Projects' permission",
		[]byte(PostAlmIntegrationsImportGitlabProjectInputSchema),
	)
}

// PostAlmIntegrationsImportGitlabProjectHandler is the handler function for the PostAlmIntegrationsImportGitlabProject tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostAlmIntegrationsImportGitlabProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/alm_integrations/import_gitlab_project", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostAlmIntegrationsImportGitlabProject")
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
