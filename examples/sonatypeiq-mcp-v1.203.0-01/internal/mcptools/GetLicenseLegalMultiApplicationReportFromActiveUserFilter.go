package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetLicenseLegalMultiApplicationReportFromActiveUserFilter tool
const GetLicenseLegalMultiApplicationReportFromActiveUserFilterInputSchema = "{\n  \"properties\": {\n    \"templateId\": {\n      \"description\": \"Enter the templateId for the license legal data.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"templateId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetLicenseLegalMultiApplicationReportFromActiveUserFilter tool (Status: 200, Content-Type: text/html)
const GetLicenseLegalMultiApplicationReportFromActiveUserFilterResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** text/html\n\n> The response contains license legal data in HTML format based on the specified template.\n\n## Response Structure\n\n- Structure (Type: string):\n"

// NewGetLicenseLegalMultiApplicationReportFromActiveUserFilterMCPTool creates the MCP Tool instance for GetLicenseLegalMultiApplicationReportFromActiveUserFilter
func NewGetLicenseLegalMultiApplicationReportFromActiveUserFilterMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLicenseLegalMultiApplicationReportFromActiveUserFilter",
		"Use this method to generate license legal data in HTML format for applications ,on which the logged in user has permissions.\n\nPermission required: Review Legal Obligations For Components Licenses",
		[]byte(GetLicenseLegalMultiApplicationReportFromActiveUserFilterInputSchema),
	)
}

// GetLicenseLegalMultiApplicationReportFromActiveUserFilterHandler is the handler function for the GetLicenseLegalMultiApplicationReportFromActiveUserFilter tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLicenseLegalMultiApplicationReportFromActiveUserFilterHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/licenseLegalMetadata/multiApplication/activeUserFilter/report/templateId/{templateId}", args, []string{"templateId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLicenseLegalMultiApplicationReportFromActiveUserFilter")
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
