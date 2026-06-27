package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the DownloadBackupFile tool
const DownloadBackupFileInputSchema = "{\n  \"properties\": {\n    \"jobId\": {\n      \"description\": \"id of the backup/restore job\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"jobId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the DownloadBackupFile tool (Status: 200, Content-Type: application/json)
const DownloadBackupFileResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a data stream of the backup file content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the DownloadBackupFile tool (Status: 404, Content-Type: application/json)
const DownloadBackupFileResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if job not found or user doesn't have permissions for the job or the file is missing\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDownloadBackupFileMCPTool creates the MCP Tool instance for DownloadBackupFile
func NewDownloadBackupFileMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DownloadBackupFile",
		"Download backup file - Downloads the backup file for the given job. Requires site admin or space export permissions for all spaces included in the backup job.",
		[]byte(DownloadBackupFileInputSchema),
	)
}

// DownloadBackupFileHandler is the handler function for the DownloadBackupFile tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DownloadBackupFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/backup-restore/jobs/{jobId}/download", args, []string{"jobId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DownloadBackupFile"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
