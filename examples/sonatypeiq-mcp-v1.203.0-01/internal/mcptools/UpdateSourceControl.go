package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the UpdateSourceControl tool
const UpdateSourceControlInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify the SCM settings for the ownerId specified above in the request JSON.\\u003cul\\u003e\\u003cli\\u003e\\u003ccode\\u003eid\\u003c/code\\u003e is the internal owner ID.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003erepositoryUrl\\u003c/code\\u003e is the http(s) and ssh urls for the application specified in the ownerId.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003eusername\\u003c/code\\u003e is optional, can be provided for Bitbucket Server and Cloud.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003etoken\\u003c/code\\u003e is optional if inherited. If provided, this value will override the value inherited from the root organization, organization or application level.\\u003cli\\u003e\\u003ccode\\u003eprovider\\u003c/code\\u003e is the name of of the SCM system. Allowed values are \\u003ccode\\u003eazure\\u003c/code\\u003e, \\u003ccode\\u003egithub\\u003c/code\\u003e, \\u003ccode\\u003egitlab\\u003c/code\\u003e, and \\u003ccode\\u003ebitbucket\\u003c/code\\u003e.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003ebaseBranch\\u003c/code\\u003e is required for the root organization. Organizations and applications inherit from the root unless overridden.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003eenablePullRequests\\u003c/code\\u003e has been deprecated in version 124.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003eremediationPullRequestsEnabled\\u003c/code\\u003e is optional. Set it to " + "\x60" + "true" + "\x60" + " to enable the Automated Pull Requests.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003eenableStatusChecks\\u003c/code\\u003e has been deprecated in version 124.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003estatusChecksEnabled\\u003c/code\\u003e is an internal field.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003epullRequestCommentingEnabled\\u003c/code\\u003e is optional. Set it to " + "\x60" + "true" + "\x60" + " to enable the  Pull Request Commenting feature.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003esourceControlEvaluationsEnabled\\u003c/code\\u003e is set to " + "\x60" + "true" + "\x60" + " to enable source control evaluations for the continuous risk profile feature.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003esourceControlScanTarget\\u003c/code\\u003e is the path inside the repository.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003esshEnabled\\u003c/code\\u003e is set to " + "\x60" + "true" + "\x60" + " to enable ssh.\\u003c/li\\u003e\\u003cli\\u003e\\u003ccode\\u003ecommitStatusEnabled\\u003c/code\\u003e is set to " + "\x60" + "true" + "\x60" + " if interaction with the commit statuses on the SCM is enabled.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"authenticationType\": {\n          \"type\": \"string\"\n        },\n        \"baseBranch\": {\n          \"type\": \"string\"\n        },\n        \"closePrAfterDays\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"closePrAfterDaysOpenEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"closePrOnFailedChecksEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"commitStatusEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"enablePullRequests\": {\n          \"type\": \"boolean\"\n        },\n        \"enableStatusChecks\": {\n          \"type\": \"boolean\"\n        },\n        \"githubAppId\": {\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"innerSourceAutomatedUpdatesEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"manualPullRequestsEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"nonGoldenPullRequestsEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"ownerId\": {\n          \"type\": \"string\"\n        },\n        \"provider\": {\n          \"type\": \"string\"\n        },\n        \"pullRequestCommentingEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"remediationPullRequestsEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"repositoryUrl\": {\n          \"type\": \"string\"\n        },\n        \"sourceControlEvaluationsEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"sourceControlScanTarget\": {\n          \"type\": \"string\"\n        },\n        \"sshEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"statusChecksEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"token\": {\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Enter the internal ownerId. Use ROOT_ORGANIZATION_ID for the root organization.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the value for ownerType.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateSourceControl tool (Status: 200, Content-Type: application/json)
const UpdateSourceControlResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The SCM settings have been updated successfully. The JSON returned shows the updated values.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **nonGoldenPullRequestsEnabled** (Type: boolean):\n  - **authenticationType** (Type: string):\n  - **sourceControlEvaluationsEnabled** (Type: boolean):\n  - **closePrOnFailedChecksEnabled** (Type: boolean):\n  - **username** (Type: string):\n  - **sshEnabled** (Type: boolean):\n  - **enableStatusChecks** (Type: boolean):\n  - **githubAppId** (Type: string):\n  - **sourceControlScanTarget** (Type: string):\n  - **token** (Type: string):\n  - **commitStatusEnabled** (Type: boolean):\n  - **enablePullRequests** (Type: boolean):\n  - **innerSourceAutomatedUpdatesEnabled** (Type: boolean):\n  - **closePrAfterDaysOpenEnabled** (Type: boolean):\n  - **statusChecksEnabled** (Type: boolean):\n  - **ownerId** (Type: string):\n  - **id** (Type: string):\n  - **pullRequestCommentingEnabled** (Type: boolean):\n  - **manualPullRequestsEnabled** (Type: boolean):\n  - **provider** (Type: string):\n  - **remediationPullRequestsEnabled** (Type: boolean):\n  - **closePrAfterDays** (Type: integer, int32):\n  - **repositoryUrl** (Type: string):\n  - **baseBranch** (Type: string):\n"

// NewUpdateSourceControlMCPTool creates the MCP Tool instance for UpdateSourceControl
func NewUpdateSourceControlMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateSourceControl",
		"Use this method to update an existing SCM setting.\n\nPermissions required: Edit IQ Elements",
		[]byte(UpdateSourceControlInputSchema),
	)
}

// UpdateSourceControlHandler is the handler function for the UpdateSourceControl tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateSourceControlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/sourceControl/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateSourceControl")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
