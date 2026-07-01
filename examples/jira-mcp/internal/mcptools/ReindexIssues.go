package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the ReindexIssues tool
const ReindexIssuesInputSchema = "{\n  \"properties\": {\n    \"indexChangeHistory\": {\n      \"default\": false,\n      \"description\": \"Indicates that changeHistory should also be reindexed.\",\n      \"type\": \"boolean\"\n    },\n    \"indexComments\": {\n      \"default\": false,\n      \"description\": \"Indicates that comments should also be reindexed.\",\n      \"type\": \"boolean\"\n    },\n    \"indexWorklogs\": {\n      \"default\": false,\n      \"description\": \"Indicates that worklogs should also be reindexed.\",\n      \"type\": \"boolean\"\n    },\n    \"issueId\": {\n      \"description\": \"The IDs or keys of one or more issues to reindex.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the ReindexIssues tool (Status: 200, Content-Type: application/json)
const ReindexIssuesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns response indicating reindex time.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **finishTime** (Type: string, date-time):\n  - **progressUrl** (Type: string):\n      - Example: 'http://localhost:8080/jira'\n  - **startTime** (Type: string, date-time):\n  - **submittedTime** (Type: string, date-time):\n  - **success** (Type: boolean):\n      - Example: 'true'\n  - **type** (Type: string):\n      - Example: 'FOREGROUND'\n      - Enum: ['FOREGROUND', 'BACKGROUND', 'BACKGROUND_PREFFERED', 'BACKGROUND_PREFERRED']\n  - **currentProgress** (Type: integer, int64):\n      - Example: '0'\n  - **currentSubTask** (Type: string):\n      - Example: 'Currently reindexing Change History'\n"

// NewReindexIssuesMCPTool creates the MCP Tool instance for ReindexIssues
func NewReindexIssuesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ReindexIssues",
		"Reindex individual issues - Reindexes one or more individual issues. Indexing is performed synchronously - the call returns when indexing of the issues has completed or a failure occurs.",
		[]byte(ReindexIssuesInputSchema),
	)
}

// ReindexIssuesHandler is the handler function for the ReindexIssues tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ReindexIssuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/reindex/issue", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ReindexIssues")
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
