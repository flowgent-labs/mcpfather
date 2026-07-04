package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostAlmIntegrationsImportBitbucketserverProject tool
const PostAlmIntegrationsImportBitbucketserverProjectInputSchema = "{\n  \"properties\": {\n    \"almSetting\": {\n      \"description\": \"DevOps Platform configuration key. This parameter is optional if you have only one BitBucket Server integration.\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    },\n    \"newCodeDefinitionType\": {\n      \"description\": \"Project New Code Definition Type New code definitions of the following types are allowed: - PREVIOUS_VERSION - NUMBER_OF_DAYS - REFERENCE_BRANCH - will default to the main branch.\",\n      \"type\": \"string\"\n    },\n    \"newCodeDefinitionValue\": {\n      \"description\": \"Project New Code Definition Value For each new code definition type, a different value is expected: - no value, when the new code definition type is PREVIOUS_VERSION and REFERENCE_BRANCH - a number between 1 and 90, when the new code definition type is NUMBER_OF_DAYS\",\n      \"type\": \"string\"\n    },\n    \"projectKey\": {\n      \"description\": \"BitbucketServer project key\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    },\n    \"repositorySlug\": {\n      \"description\": \"BitbucketServer repository slug\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectKey\",\n    \"repositorySlug\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostAlmIntegrationsImportBitbucketserverProjectMCPTool creates the MCP Tool instance for PostAlmIntegrationsImportBitbucketserverProject
func NewPostAlmIntegrationsImportBitbucketserverProjectMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostAlmIntegrationsImportBitbucketserverProject",
		"Create a SonarQube project with the information from the provided BitbucketServer project. - Create a SonarQube project with the information from the provided BitbucketServer project. Autoconfigure pull request decoration mechanism. Requires the 'Create Projects' permission",
		[]byte(PostAlmIntegrationsImportBitbucketserverProjectInputSchema),
	)
}

// PostAlmIntegrationsImportBitbucketserverProjectHandler is the handler function for the PostAlmIntegrationsImportBitbucketserverProject tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostAlmIntegrationsImportBitbucketserverProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/alm_integrations/import_bitbucketserver_project", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostAlmIntegrationsImportBitbucketserverProject")
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
