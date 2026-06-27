package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the CreateSpaceRestoreJob tool
const CreateSpaceRestoreJobInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"space restore settings\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the CreateSpaceRestoreJob tool (Status: 200, Content-Type: application/json)
const CreateSpaceRestoreJobResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the space restore job.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateSpaceRestoreJob tool (Status: 400, Content-Type: application/json)
const CreateSpaceRestoreJobResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  Returned if invalid filename provided\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateSpaceRestoreJob tool (Status: 403, Content-Type: application/json)
const CreateSpaceRestoreJobResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if user doesn't have permission to restore spaces\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreateSpaceRestoreJobMCPTool creates the MCP Tool instance for CreateSpaceRestoreJob
func NewCreateSpaceRestoreJobMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateSpaceRestoreJob",
		"Create space restore job - Creates new space restore job and adds it to the queue.",
		[]byte(CreateSpaceRestoreJobInputSchema),
	)
}

// CreateSpaceRestoreJobHandler is the handler function for the CreateSpaceRestoreJob tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateSpaceRestoreJobHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/backup-restore/restore/space", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateSpaceRestoreJob"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
