package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetCompositeSourceControlByOwner tool
const GetCompositeSourceControlByOwnerInputSchema = "{\n  \"properties\": {\n    \"internalOwnerId\": {\n      \"description\": \"Enter the id of the application or organization for which you want to retrieve the composite source control configuration settings\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the ownerType of the entity (organization or application) for which you want to retrieve the composite source control configuration settings.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetCompositeSourceControlByOwner tool (Status: 200, Content-Type: application/json)
const GetCompositeSourceControlByOwnerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains values for the SCM configuration. For each value, the corresponding parent value will be shown, if applicable.<ul><li><code>id</code> is the internal identifier for the SCM configuration.</li><li><code>ownerId</code> is the identifier for the ownerType specified.</li><li><code>repositoryUrl</code> indicates the URL of application/organization. Will indicate 'null' for organizations.</li><li><code>provider</code> is the name of the source code host for the parent. Values can be Azure, GitHub, GitLab and Bitbucket.</li><li><code>username</code> is returned if found for the specific provider. Currently, the values are available for Bitbucket Server and Bitbucket Cloud.</li><li><code>token</code> is obfuscated and indicates the composite configuration for the source control host.<li><code>baseBranch</code> shows the base branch name.<li><code>remediationPullRequestsEnabled</code> indicates if the Automated Pull Request feature is enabled.</li><li><code>statusChecksEnabled</code> indicates if the status checks for the source code are enabled.</li><li><code>pullRequestCommentingEnabled</code> indicates if PR commenting is enabled for this application/organization.</li><li><code>sourceControlEvaluationsEnabled</code> indicates if the evaluations triggered by the IQ Server are enabled, for the Continuous Risk Profile feature.</li><li><code>sshEnabled</code> indicates if ssh settings are enabled.</li><li><code>commitStatusEnabled</code> indicates if commit status check is enabled.</li>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **sourceControlScanTarget** (Type: object):\n    - **value** (Type: string):\n    - **parentName** (Type: string):\n    - **parentValue** (Type: string):\n  - **[cyclic reference]**\n  - **closePrOnFailedChecksEnabled** (Type: object):\n    - **parentValue** (Type: boolean):\n    - **value** (Type: boolean):\n    - **parentName** (Type: string):\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **githubApps** (Type: array):\n    - **Items** (Type: object):\n      - **parentValue** (Type: object):\n        - **accountName** (Type: string):\n        - **configurationDate** (Type: string):\n        - **id** (Type: string):\n        - **installationId** (Type: integer, int64):\n        - **isActive** (Type: boolean):\n        - **name** (Type: string):\n      - **[cyclic reference]**\n      - **parentName** (Type: string):\n  - **id** (Type: string):\n  - **ownerId** (Type: string):\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **closePrAfterDays** (Type: object):\n    - **parentName** (Type: string):\n    - **parentValue** (Type: integer, int32):\n    - **value** (Type: integer, int32):\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **repositoryUrl** (Type: string):\n"

// NewGetCompositeSourceControlByOwnerMCPTool creates the MCP Tool instance for GetCompositeSourceControlByOwner
func NewGetCompositeSourceControlByOwnerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCompositeSourceControlByOwner",
		"Use this method to retrieve the composite source control management (SCM) configuration settings.\n\nPermissions required: View IQ Elements",
		[]byte(GetCompositeSourceControlByOwnerInputSchema),
	)
}

// GetCompositeSourceControlByOwnerHandler is the handler function for the GetCompositeSourceControlByOwner tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCompositeSourceControlByOwnerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/compositeSourceControl/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetCompositeSourceControlByOwner")
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
