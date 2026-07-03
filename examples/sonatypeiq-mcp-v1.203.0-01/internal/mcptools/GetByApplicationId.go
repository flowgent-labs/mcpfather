package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetByApplicationId tool
const GetByApplicationIdInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal application Id. You can use the Applications REST API to get the internal application Id. \",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetByApplicationId tool (Status: 200, Content-Type: application/json)
const GetByApplicationIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response JSON contains the URLs to access the latest scan report for the applicationId provided. \n\nThe response field " + "\x60" + "stage" + "\x60" + " indicates the stage at which the policy evaluation was executed, such as 'develop', 'build', 'release'.  The response field " + "\x60" + "latestReportHtmlUrl" + "\x60" + " is a relative link to view the most recent report. Response fields " + "\x60" + "reportPdfURL" + "\x60" + " and " + "\x60" + "reportHtmlURL" + "\x60" + " are links to view the pdf version of the report. The response field " + "\x60" + "reportDataUrl" + "\x60" + " is a link to view the most recent report data. \n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **embeddableReportHtmlUrl** (Type: string):\n    - **evaluationDate** (Type: string, date-time):\n    - **latestReportHtmlUrl** (Type: string):\n    - **reportDataUrl** (Type: string):\n    - **reportHtmlUrl** (Type: string):\n    - **reportPdfUrl** (Type: string):\n    - **stage** (Type: string):\n    - **applicationId** (Type: string):\n"

// NewGetByApplicationIdMCPTool creates the MCP Tool instance for GetByApplicationId
func NewGetByApplicationIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetByApplicationId",
		"Use this method to retrieve the application reports for the specified application Id. You can view application reports only for applications to which you have access. \n\nPermissions required: View IQ Elements ",
		[]byte(GetByApplicationIdInputSchema),
	)
}

// GetByApplicationIdHandler is the handler function for the GetByApplicationId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetByApplicationIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/reports/applications/{applicationId}", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetByApplicationId")
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
