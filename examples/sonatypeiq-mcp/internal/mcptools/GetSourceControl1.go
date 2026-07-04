package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetSourceControl1 tool
const GetSourceControl1InputSchema = "{\n  \"properties\": {\n    \"internalOwnerId\": {\n      \"description\": \"Enter the value for internal ownerId. Use ROOT_ORGANIZATION_ID for the root organization\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the value for ownerType.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetSourceControl1 tool (Status: 200, Content-Type: application/json)
const GetSourceControl1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains source control configuration settings for the specified ownerId.\n\n<ul><li><code>id</code> is the owner internal ID.</li><li><code>repositoryUrl</code> indicates the http(s) and ssh urls for the application specified in the ownerId.</li><li><code>username</code> is retrieved if available on the SCM system, e.g. for Bitbucket Server and Cloud.</li><li><code>provider</code> indicates the name of the SCM system.</li><li><code>baseBranch</code> indicates the name of the last selected branch.</li><li><code>enablePullRequests</code> has been deprecated in version 124.</li><li><code>remediationPullRequestsEnabled</code> indicates if the Automated Pull Requests feature is enabled.</li><li><code>enableStatusChecks</code> has been deprecated in version 124.</li><li><code>statusChecksEnabled</code> is an internal field.</li><li><code>pullRequestCommentingEnabled</code> indicates if the Pull Request Commenting feature is enabled.</li><li><code>sourceControlEvaluationsEnabled</code> indicates if the source control evaluations are enabled for the continuous risk profile feature.</li><li><code>sourceControlScanTarget</code> indicates the path inside the repository.</li><li><code>sshEnabled</code> indicates if ssh is enabled.</li><li><code>commitStatusEnabled</code> indicates if interaction with the commit statuses on the SCM system is enabled.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **provider** (Type: string):\n  - **closePrAfterDaysOpenEnabled** (Type: boolean):\n  - **authenticationType** (Type: string):\n  - **nonGoldenPullRequestsEnabled** (Type: boolean):\n  - **closePrAfterDays** (Type: integer, int32):\n  - **repositoryUrl** (Type: string):\n  - **sourceControlScanTarget** (Type: string):\n  - **remediationPullRequestsEnabled** (Type: boolean):\n  - **username** (Type: string):\n  - **manualPullRequestsEnabled** (Type: boolean):\n  - **ownerId** (Type: string):\n  - **baseBranch** (Type: string):\n  - **closePrOnFailedChecksEnabled** (Type: boolean):\n  - **enableStatusChecks** (Type: boolean):\n  - **pullRequestCommentingEnabled** (Type: boolean):\n  - **id** (Type: string):\n  - **enablePullRequests** (Type: boolean):\n  - **token** (Type: string):\n  - **commitStatusEnabled** (Type: boolean):\n  - **statusChecksEnabled** (Type: boolean):\n  - **sshEnabled** (Type: boolean):\n  - **sourceControlEvaluationsEnabled** (Type: boolean):\n  - **githubAppId** (Type: string):\n  - **innerSourceAutomatedUpdatesEnabled** (Type: boolean):\n"

// NewGetSourceControl1MCPTool creates the MCP Tool instance for GetSourceControl1
func NewGetSourceControl1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSourceControl1",
		"Use this method to retrieve the source control configuration settings for an organization or an application.\n\nPermissions required: View IQ Elements",
		[]byte(GetSourceControl1InputSchema),
	)
}

// GetSourceControl1Handler is the handler function for the GetSourceControl1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSourceControl1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/sourceControl/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSourceControl1")
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
