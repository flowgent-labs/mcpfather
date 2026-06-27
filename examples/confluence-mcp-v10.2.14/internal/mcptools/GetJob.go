package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetJob tool
const GetJobInputSchema = "{\n  \"properties\": {\n    \"jobId\": {\n      \"description\": \"id of the backup/restore job\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"jobId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetJob tool (Status: 200, Content-Type: application/json)
const GetJobResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the backup/restore job\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetJob tool (Status: 400, Content-Type: application/json)
const GetJobResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  Returned if jobId is null\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetJob tool (Status: 403, Content-Type: application/json)
const GetJobResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n>  Returned if job not found or user doesn't have permission to see it\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetJobMCPTool creates the MCP Tool instance for GetJob
func NewGetJobMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetJob",
		"Get job by ID - Get job by id. The user must be a sysadmin or the owner of the job.",
		[]byte(GetJobInputSchema),
	)
}

// GetJobHandler is the handler function for the GetJob tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetJobHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/backup-restore/jobs/{jobId}", args, []string{"jobId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetJob"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
