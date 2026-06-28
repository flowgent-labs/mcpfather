package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetLicenseLegalApplicationHtmlreport tool
const GetLicenseLegalApplicationHtmlreportInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the application id or public id.\",\n      \"type\": \"string\"\n    },\n    \"stageId\": {\n      \"description\": \"Enter the stageId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"stageId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetLicenseLegalApplicationHtmlreport tool (Status: 200, Content-Type: text/html)
const GetLicenseLegalApplicationHtmlreportResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** text/html\n\n> The response contains license legal data in HTML format.\n\n## Response Structure\n\n- Structure (Type: string):\n"

// NewGetLicenseLegalApplicationHtmlreportMCPTool creates the MCP Tool instance for GetLicenseLegalApplicationHtmlreport
func NewGetLicenseLegalApplicationHtmlreportMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLicenseLegalApplicationHtmlreport",
		"Use this method to retrieve the license legal data for components in an application at a specific stage, in HTML format.\n\nPermissions required: Review Legal Obligations For Components Licenses",
		[]byte(GetLicenseLegalApplicationHtmlreportInputSchema),
	)
}

// GetLicenseLegalApplicationHtmlreportHandler is the handler function for the GetLicenseLegalApplicationHtmlreport tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLicenseLegalApplicationHtmlreportHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/licenseLegalMetadata/application/{applicationId}/stage/{stageId}/report", args, []string{"applicationId", "stageId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLicenseLegalApplicationHtmlreport")
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
