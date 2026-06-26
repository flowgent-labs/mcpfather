package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the CreateSpaceBackupJob tool
const CreateSpaceBackupJobInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Space backup settings\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the CreateSpaceBackupJob tool (Status: 200, Content-Type: application/json)
const CreateSpaceBackupJobResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the space backup job\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateSpaceBackupJob tool (Status: 400, Content-Type: application/json)
const CreateSpaceBackupJobResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if invalid settings provided\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateSpaceBackupJob tool (Status: 403, Content-Type: application/json)
const CreateSpaceBackupJobResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if user doesn't have permission to create space backups\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateSpaceBackupJob tool (Status: 409, Content-Type: application/json)
const CreateSpaceBackupJobResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Returned if backup with the same spaces selected is already in PROGRESS or QUEUED\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreateSpaceBackupJobMCPTool creates the MCP Tool instance for CreateSpaceBackupJob
func NewCreateSpaceBackupJobMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateSpaceBackupJob",
		"Create space backup job - Creates new space backup job and adds it to the queue.",
		[]byte(CreateSpaceBackupJobInputSchema),
	)
}

// CreateSpaceBackupJobHandler is the handler function for the CreateSpaceBackupJob tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateSpaceBackupJobHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/backup-restore/backup/space", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateSpaceBackupJob"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
