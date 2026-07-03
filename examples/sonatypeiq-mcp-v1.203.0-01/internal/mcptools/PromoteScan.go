package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the PromoteScan tool
const PromoteScanInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal applicationId. Use the Applications REST API to retrieve the internal applicationId.\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"You can provide either the scanId (reportId) of the previous scan OR the source stageId (possible values are 'build', 'stage-release', 'release' or 'operate'). When using the stageId, the latest scanId for the application will be used. Enter the targetStageId for the new stage you want your scan to be promoted to (possible values are 'build', 'stage-release', 'release' or 'operate'). Using the same value for source and target stage will resubmit the latest scan report.\",\n      \"properties\": {\n        \"scanId\": {\n          \"type\": \"string\"\n        },\n        \"sourceStageId\": {\n          \"type\": \"string\"\n        },\n        \"targetStageId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PromoteScan tool (Status: 200, Content-Type: application/json)
const PromoteScanResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response will contain the statusUrl to view the evaluation result using the GET method (step 2)\n\n## Response Structure\n\n- Structure (Type: object):\n  - **statusUrl** (Type: string):\n"

// NewPromoteScanMCPTool creates the MCP Tool instance for PromoteScan
func NewPromoteScanMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PromoteScan",
		"Use this method to rescan older scans. This is done when the binaries of a previous build are now moving to a new stage in the production pipeline. Using this method, you can avoid rebuilding the application and reuse the scan metadata at the newer stage. This new evaluation will evaluate the most recent security and license data against your current policies. \n\nPermissions Required: Evaluate Applications",
		[]byte(PromoteScanInputSchema),
	)
}

// PromoteScanHandler is the handler function for the PromoteScan tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PromoteScanHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/evaluation/applications/{applicationId}/promoteScan", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PromoteScan")
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
