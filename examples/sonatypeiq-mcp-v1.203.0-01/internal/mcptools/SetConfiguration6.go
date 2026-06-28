package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetConfiguration6 tool
const SetConfiguration6InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Provide the settings for the SCM configuration as below: \\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "cloneDirectory" + "\x60" + " is the location of the cloned repository that will be used by the IQ server. If a relative path is provided, then that path will be created inside the  " + "\x60" + "sonatype-work" + "\x60" + " directory and your repository will be created within this. A return value " + "\x60" + "source-control" + "\x60" + " indicates that this setting is not configured.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "gitImplementation" + "\x60" + " will have the value " + "\x60" + "java" + "\x60" + " for JGit or " + "\x60" + "native" + "\x60" + " for a native git client.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "prCommentPurgeWindow" + "\x60" + " is the number of days until the comments of a Pull Request (PR) are allowed to be purged.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "prEventPurgeWindow" + "\x60" + " is the number of days until PR events are allowed to be purged.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "gitExecutable" + "\x60" + " is the absolute path to a native client. No value indicates the native git client is on the system path.\\u003c/li\\u003e" + "\x60" + "gitTimeoutSeconds" + "\x60" + " is the number of seconds a git command can execute before timing out.\\u003c/li\\u003e" + "\x60" + "commitUsername" + "\x60" + " is the username that will be used for the SCM features. The value " + "\x60" + "NexusIQ" + "\x60" + " indicates the default value.\\u003c/li\\u003e" + "\x60" + "commitEmail" + "\x60" + " is the commit email that will be used for the SCM features." + "\x60" + "useUsernameInRepositoryCloneUrl" + "\x60" + " indicates if the username will be added to the URL for the clonedrepository. This can be used in conjunction with " + "\x60" + "commitEmail" + "\x60" + " to support the 'Verified Committer' feature of Bitbucket.\\u003c/li\\u003e" + "\x60" + "defaultBranchMonitoringStartTime" + "\x60" + " has a default value between 00:00 and 00:10. It is the time at which the default branch monitoring will start for the first time.\\u003c/li\\u003e" + "\x60" + "defaultBranchMonitoringIntervalHours" + "\x60" + " is the number of hours elapsed between the executions of default branch monitoring by the IQ Server. The default value is 24 hours.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "pullRequestMonitoringIntervalSeconds" + "\x60" + " is the time in seconds between consecutive execution of PR monitoring. The default value is 60 seconds.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"cloneDirectory\": {\n          \"type\": \"string\"\n        },\n        \"commitEmail\": {\n          \"type\": \"string\"\n        },\n        \"commitUsername\": {\n          \"type\": \"string\"\n        },\n        \"defaultBranchMonitoringIntervalHours\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"defaultBranchMonitoringStartTime\": {\n          \"type\": \"string\"\n        },\n        \"gitExecutable\": {\n          \"type\": \"string\"\n        },\n        \"gitImplementation\": {\n          \"enum\": [\n            \"native\",\n            \"java\"\n          ],\n          \"type\": \"string\"\n        },\n        \"gitTimeoutSeconds\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"gpgPassphrase\": {\n          \"type\": \"string\"\n        },\n        \"gpgSigningKey\": {\n          \"type\": \"string\"\n        },\n        \"prCommentPurgeWindow\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"prEventPurgeWindow\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"pullRequestMonitoringIntervalSeconds\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"useUsernameInRepositoryCloneUrl\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewSetConfiguration6MCPTool creates the MCP Tool instance for SetConfiguration6
func NewSetConfiguration6MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetConfiguration6",
		"Use this method to set an SCM Configuration with the IQ Server.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(SetConfiguration6InputSchema),
	)
}

// SetConfiguration6Handler is the handler function for the SetConfiguration6 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetConfiguration6Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/sourceControl", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetConfiguration6")
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
