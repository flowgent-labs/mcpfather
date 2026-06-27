package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the ScheduleUserAnonymizationRerun tool
const ScheduleUserAnonymizationRerunInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"JSON containing parameters to schedule the anonymization process with\",\n      \"properties\": {\n        \"newOwnerKey\": {\n          \"example\": \"admin\",\n          \"type\": \"string\"\n        },\n        \"oldUserKey\": {\n          \"type\": \"string\"\n        },\n        \"oldUserName\": {\n          \"type\": \"string\"\n        },\n        \"userKey\": {\n          \"example\": \"JIRAUSER10100\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// NewScheduleUserAnonymizationRerunMCPTool creates the MCP Tool instance for ScheduleUserAnonymizationRerun
func NewScheduleUserAnonymizationRerunMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ScheduleUserAnonymizationRerun",
		"Schedule user anonymization rerun - Schedules a user anonymization process. Requires system admin permission.",
		[]byte(ScheduleUserAnonymizationRerunInputSchema),
	)
}

// ScheduleUserAnonymizationRerunHandler is the handler function for the ScheduleUserAnonymizationRerun tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ScheduleUserAnonymizationRerunHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/user/anonymization/rerun", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ScheduleUserAnonymizationRerun"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
