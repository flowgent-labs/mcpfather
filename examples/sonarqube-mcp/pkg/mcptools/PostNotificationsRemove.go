package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostNotificationsRemove tool
const PostNotificationsRemoveInputSchema = "{\n  \"properties\": {\n    \"channel\": {\n      \"description\": \"Channel through which the notification is sent. For example, notifications can be sent by email.\",\n      \"enum\": [\n        \"EmailNotificationChannel\"\n      ],\n      \"type\": \"string\"\n    },\n    \"login\": {\n      \"description\": \"User login\",\n      \"type\": \"string\"\n    },\n    \"project\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"description\": \"Notification type. Possible values are for: - Global notifications: CeReportTaskFailure, ChangesOnMyIssue, NewAlerts, QualityGateConditionsMismatch, SQ-MyNewIssues - Per project notifications: CeReportTaskFailure, ChangesOnMyIssue, NewAlerts, NewFalsePositiveIssue, NewIssues, SQ-MyNewIssues\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostNotificationsRemoveMCPTool creates the MCP Tool instance for PostNotificationsRemove
func NewPostNotificationsRemoveMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostNotificationsRemove",
		"Remove a notification for the authenticated user. - Remove a notification for the authenticated user. Requires one of the following permissions: - Authentication if no login is provided - System administration if a login is provided",
		[]byte(PostNotificationsRemoveInputSchema),
	)
}

// PostNotificationsRemoveHandler is the handler function for the PostNotificationsRemove tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostNotificationsRemoveHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/notifications/remove", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostNotificationsRemove")
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
