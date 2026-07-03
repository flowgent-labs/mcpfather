package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration6 tool
const GetConfiguration6InputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration6 tool (Status: 200, Content-Type: application/json)
const GetConfiguration6ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains: <ul><li>" + "\x60" + "cloneDirectory" + "\x60" + " is the location of the cloned repository that will be used by the IQ server. If a relative path is provided, then that path will be created inside the  " + "\x60" + "sonatype-work" + "\x60" + " directory and your repository will be created within this. A return value " + "\x60" + "source-control" + "\x60" + " indicates that this setting is not configured.</li><li>" + "\x60" + "gitImplementation" + "\x60" + " will have the value " + "\x60" + "java" + "\x60" + " for JGit or " + "\x60" + "native" + "\x60" + " for a native git client.</li><li>" + "\x60" + "prCommentPurgeWindow" + "\x60" + " is the number of days until the comments of a Pull Request (PR) are allowed to be purged.</li><li>" + "\x60" + "prEventPurgeWindow" + "\x60" + " is the number of days until PR events are allowed to be purged.</li><li>" + "\x60" + "gitExecutable" + "\x60" + " is the absolute path to a native client. No value indicates the native git client is on the system path.</li>" + "\x60" + "gitTimeoutSeconds" + "\x60" + " is the number of seconds a git command can execute before timing out.</li>" + "\x60" + "commitUsername" + "\x60" + " is the username that will be used for the SCM features. The value " + "\x60" + "NexusIQ" + "\x60" + " indicates the default value.</li>" + "\x60" + "commitEmail" + "\x60" + " is the commit email that will be used for the SCM features." + "\x60" + "useUsernameInRepositoryCloneUrl" + "\x60" + " indicates if the username will be added to the URL for the cloned repository. This can be used in conjunction with " + "\x60" + "commitEmail" + "\x60" + " to support the  'Verified Committer' feature of Bitbucket.</li>" + "\x60" + "defaultBranchMonitoringStartTime" + "\x60" + " has a default value between 00:00 and 00:10. It is the time at which the default branch monitoring will start for the first time.</li>" + "\x60" + "defaultBranchMonitoringIntervalHours" + "\x60" + " is the number of hours elapsed between the executions of default branch monitoring by the IQ Server. The default value is 24 hours.</li><li>" + "\x60" + "pullRequestMonitoringIntervalSeconds" + "\x60" + " is the time in seconds between consecutive execution of PR monitoring. The default value is 60 seconds.</li></ul> \n\n## Response Structure\n\n- Structure (Type: object):\n  - **commitUsername** (Type: string):\n  - **defaultBranchMonitoringStartTime** (Type: string):\n  - **prEventPurgeWindow** (Type: integer, int32):\n  - **cloneDirectory** (Type: string):\n  - **gitExecutable** (Type: string):\n  - **gitImplementation** (Type: string):\n      - Enum: ['native', 'java']\n  - **gpgPassphrase** (Type: string):\n  - **prCommentPurgeWindow** (Type: integer, int32):\n  - **defaultBranchMonitoringIntervalHours** (Type: integer, int32):\n  - **gpgSigningKey** (Type: string):\n  - **gitTimeoutSeconds** (Type: integer, int32):\n  - **pullRequestMonitoringIntervalSeconds** (Type: integer, int32):\n  - **useUsernameInRepositoryCloneUrl** (Type: boolean):\n  - **commitEmail** (Type: string):\n"

// NewGetConfiguration6MCPTool creates the MCP Tool instance for GetConfiguration6
func NewGetConfiguration6MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration6",
		"Use this method to retrieve an existing SCM configuration.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetConfiguration6InputSchema),
	)
}

// GetConfiguration6Handler is the handler function for the GetConfiguration6 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguration6Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/sourceControl", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration6")
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
