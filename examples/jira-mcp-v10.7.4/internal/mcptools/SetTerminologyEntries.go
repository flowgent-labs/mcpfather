package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetTerminologyEntries tool
const SetTerminologyEntriesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Collection of TerminologyRequestBean\",\n      \"properties\": {\n        \"newName\": {\n          \"example\": \"Theme\",\n          \"type\": \"string\"\n        },\n        \"newNamePlural\": {\n          \"example\": \"Themes\",\n          \"type\": \"string\"\n        },\n        \"originalName\": {\n          \"example\": \"Epic\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetTerminologyEntriesMCPTool creates the MCP Tool instance for SetTerminologyEntries
func NewSetTerminologyEntriesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetTerminologyEntries",
		"Update epic/sprint names from original to new - Change epic/sprint names from {originalName} to {newName}. The {newName} will be displayed in Jira instead of {originalName}\n{\"originalName\"} must be equal to \"epic\" or \"sprint\".\nThere can be only one entry per unique {\"originalName\"}.\n{\"newName\"} can only consist of alphanumeric characters and spaces e.g. {\"newName\": \"iteration number 2\"}.\n{\"newName\"} must be between 1 to 100 characters.\nIt can't use the already defined {\"newName\"} values or restricted JQL words.\nTo reset {\"newName\"} to the default value, enter the {\"originalName\"} value as the value for {\"newName\"}. For example, if you want to return to {\"originalName\": \"sprint\"}, enter {\"newName\": \"sprint\"}.",
		[]byte(SetTerminologyEntriesInputSchema),
	)
}

// SetTerminologyEntriesHandler is the handler function for the SetTerminologyEntries tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetTerminologyEntriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/terminology/entries", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetTerminologyEntries"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
