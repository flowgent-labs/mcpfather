package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetApplicationEvaluationStatus tool
const GetApplicationEvaluationStatusInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the applicationId, for the which policy evaluation was requested.\",\n      \"type\": \"string\"\n    },\n    \"statusId\": {\n      \"description\": \"Enter the statusId value obtained as response of the POST call in step 1.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"statusId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetApplicationEvaluationStatus tool (Status: 200, Content-Type: application/json)
const GetApplicationEvaluationStatusResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response will include one of the 3 possible status values: PENDING (indicates that the evaluation is still in progress), FAILED or COMPLETED. For completed evaluations, the response will contain the URLs for evaluation report to view the evaluation results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **embeddableReportHtmlUrl** (Type: string):\n  - **reason** (Type: string):\n  - **reportDataUrl** (Type: string):\n  - **reportHtmlUrl** (Type: string):\n  - **reportPdfUrl** (Type: string):\n  - **status** (Type: string):\n"

// NewGetApplicationEvaluationStatusMCPTool creates the MCP Tool instance for GetApplicationEvaluationStatus
func NewGetApplicationEvaluationStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplicationEvaluationStatus",
		"This is step 2 of the policy evaluation process. Use the statusUrl obtained from the POST response for the corresponding applicationId. \n\nPermissions Required: Evaluate Applications",
		[]byte(GetApplicationEvaluationStatusInputSchema),
	)
}

// GetApplicationEvaluationStatusHandler is the handler function for the GetApplicationEvaluationStatus tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicationEvaluationStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/evaluation/applications/{applicationId}/status/{statusId}", args, []string{"applicationId", "statusId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplicationEvaluationStatus")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
