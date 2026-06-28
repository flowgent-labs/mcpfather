package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetScanStatus tool
const GetScanStatusInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the application internal id for the SBOM to be evaluated.\",\n      \"type\": \"string\"\n    },\n    \"scanRequestId\": {\n      \"description\": \"Enter the statusId from the statusUrl generated as a response to the POST request to perform the evaluation.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"scanRequestId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetScanStatus tool (Status: 200, Content-Type: application/json)
const GetScanStatusResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains summarized results of the SBOM evaluation and the URLs for detailed evaluation reports in HTML, pdf and raw formats.\n\n" + "\x60" + "policyAction" + "\x60" + " indicates the policy actions determined by the " + "\x60" + "stageId" + "\x60" + " selected while submitting the evaluation using the POST method.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **errorMessage** (Type: string):\n  - **grandfatheredPolicyViolations** (Type: integer, int32):\n  - **openPolicyViolations** (Type: object):\n    - **severe** (Type: integer, int32):\n    - **critical** (Type: integer, int32):\n    - **moderate** (Type: integer, int32):\n  - **legacyViolations** (Type: integer, int32):\n  - **policyAction** (Type: string):\n  - **reportDataUrl** (Type: string):\n  - **reportPdfUrl** (Type: string):\n  - **[cyclic reference]**\n  - **isError** (Type: boolean):\n  - **reportHtmlUrl** (Type: string):\n  - **embeddableReportHtmlUrl** (Type: string):\n"

// NewGetScanStatusMCPTool creates the MCP Tool instance for GetScanStatus
func NewGetScanStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetScanStatus",
		"SBOM evaluation is an asynchronous operation. Use this method to check on the status of the SBOM evaluation.\n\nPermissions required: Evaluate Applications",
		[]byte(GetScanStatusInputSchema),
	)
}

// GetScanStatusHandler is the handler function for the GetScanStatus tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetScanStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/scan/applications/{applicationId}/status/{scanRequestId}", args, []string{"applicationId", "scanRequestId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetScanStatus")
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
