package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAll1 tool
const GetAll1InputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAll1 tool (Status: 200, Content-Type: application/json)
const GetAll1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response JSON contains URLs to view the report data in html and pdf format, for each application to which you have access.\n\nThe response field stage indicates the stage at which the policy evaluation was executed, such as 'develop', 'build' and 'release' The response field latestReportHtmlUrl is a relative link to view the most recent report. Response fields reportPdfUrl and reportHtmlUrl are links to view the pdf version of the report.The response field reportDataUrl is a link to view the most recent report data. \n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **applicationId** (Type: string):\n    - **embeddableReportHtmlUrl** (Type: string):\n    - **evaluationDate** (Type: string, date-time):\n    - **latestReportHtmlUrl** (Type: string):\n    - **reportDataUrl** (Type: string):\n    - **reportHtmlUrl** (Type: string):\n    - **reportPdfUrl** (Type: string):\n    - **stage** (Type: string):\n"

// NewGetAll1MCPTool creates the MCP Tool instance for GetAll1
func NewGetAll1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAll1",
		"Use this method to view all application reports for applications to which  you have access. \n\nPermissions required: View IQ Elements ",
		[]byte(GetAll1InputSchema),
	)
}

// GetAll1Handler is the handler function for the GetAll1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAll1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/reports/applications", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAll1")
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
