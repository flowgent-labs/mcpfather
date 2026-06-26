package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateVersion tool
const UpdateVersionInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"JSON containing parameters to update the version with\",\n      \"properties\": {\n        \"archived\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"description\": {\n          \"example\": \"An excellent version\",\n          \"type\": \"string\"\n        },\n        \"expand\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"example\": \"10000\",\n          \"type\": \"string\"\n        },\n        \"moveUnfixedIssuesTo\": {\n          \"example\": \"http://localhost:8090/jira/rest/api/2/version/10000/move\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"New Version 1\",\n          \"type\": \"string\"\n        },\n        \"overdue\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"project\": {\n          \"example\": \"PXA\",\n          \"type\": \"string\"\n        },\n        \"projectId\": {\n          \"example\": 10000,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"releaseDate\": {\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"releaseDateSet\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"released\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"self\": {\n          \"example\": \"http://localhost:8090/jira/rest/api/2/version/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        },\n        \"startDate\": {\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"startDateSet\": {\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"userReleaseDate\": {\n          \"example\": \"2012-09-15T21:11:01.834+0000\",\n          \"type\": \"string\"\n        },\n        \"userStartDate\": {\n          \"example\": \"2012-08-15T21:11:01.834+0000\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"ID of the version.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateVersionMCPTool creates the MCP Tool instance for UpdateVersion
func NewUpdateVersionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateVersion",
		"Update version details - Updates a version.",
		[]byte(UpdateVersionInputSchema),
	)
}

// UpdateVersionHandler is the handler function for the UpdateVersion tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateVersionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/version/{id}", args, []string{"id"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateVersion"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
