package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetLicenseLegalCustomApplicationHtmlreport1 tool
const GetLicenseLegalCustomApplicationHtmlreport1InputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the application id or public id.\",\n      \"type\": \"string\"\n    },\n    \"stageId\": {\n      \"description\": \"Enter the stageId.\",\n      \"type\": \"string\"\n    },\n    \"templateId\": {\n      \"description\": \"Enter the templateId for the HTML report format.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"stageId\",\n    \"templateId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetLicenseLegalCustomApplicationHtmlreport1MCPTool creates the MCP Tool instance for GetLicenseLegalCustomApplicationHtmlreport1
func NewGetLicenseLegalCustomApplicationHtmlreport1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLicenseLegalCustomApplicationHtmlreport1",
		"Use this method to generate a license legal report in the specified HTML template format.\n\nPermissions required: Review Legal Obligations For Components Licenses",
		[]byte(GetLicenseLegalCustomApplicationHtmlreport1InputSchema),
	)
}

// GetLicenseLegalCustomApplicationHtmlreport1Handler is the handler function for the GetLicenseLegalCustomApplicationHtmlreport1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLicenseLegalCustomApplicationHtmlreport1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/licenseLegalMetadata/application/{applicationId}/stage/{stageId}/report/templateId/{templateId}", args, []string{"applicationId", "stageId", "templateId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLicenseLegalCustomApplicationHtmlreport1")
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
