package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the FindJobs tool
const FindJobsInputSchema = "{\n  \"properties\": {\n    \"fromDate\": {\n      \"description\": \"minimum job creation date. Supported date format is " + "\x60" + "yyyy-MM-ddTHH:mm:ss.SSSZ" + "\x60" + "\",\n      \"type\": \"string\"\n    },\n    \"jobOperation\": {\n      \"description\": \"job operation. Acceptable values: \\\"BACKUP\\\" and \\\"RESTORE\\\"\",\n      \"type\": \"string\"\n    },\n    \"jobScope\": {\n      \"description\": \"scope of the job. Acceptable values: \\\"SPACE\\\" and \\\"SITE\\\" \",\n      \"type\": \"string\"\n    },\n    \"jobStates\": {\n      \"description\": \"list of job states. Acceptable values: \\\"QUEUED\\\", \\\"PROCESSING\\\", \\\"FINISHED\\\", \\\"CANCELLING\\\", \\\"CANCELLED\\\", \\\"FAILED\\\"\",\n      \"items\": {\n        \"enum\": [\n          \"QUEUED\",\n          \"PROCESSING\",\n          \"COMPLETING\",\n          \"FINISHED\",\n          \"CANCELLING\",\n          \"CANCELLED\",\n          \"FAILED\"\n        ],\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"amount of jobs that should be returned\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"owner\": {\n      \"description\": \"userName of user who had created a job.\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the key of the Space the User is attempting to view.\",\n      \"type\": \"string\"\n    },\n    \"toDate\": {\n      \"description\": \"maximum job create date. Supported date format is " + "\x60" + "yyyy-MM-ddTHH:mm:ss.SSSZ" + "\x60" + "\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the FindJobs tool (Status: 200, Content-Type: application/json)
const FindJobsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the List of backup/restore jobs visible to user based on the filter provided and the user's permissions.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: unknown):\n"

// Response Template for the FindJobs tool (Status: 400, Content-Type: application/json)
const FindJobsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if invalid filter parameters were passed\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewFindJobsMCPTool creates the MCP Tool instance for FindJobs
func NewFindJobsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FindJobs",
		"Find jobs by filters - Returns jobs based on the filters provided. The user must have permission to see the jobs.",
		[]byte(FindJobsInputSchema),
	)
}

// FindJobsHandler is the handler function for the FindJobs tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FindJobsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/backup-restore/jobs", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "FindJobs"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
