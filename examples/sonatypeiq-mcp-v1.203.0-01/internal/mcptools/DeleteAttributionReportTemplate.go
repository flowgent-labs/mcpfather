package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteAttributionReportTemplate tool
const DeleteAttributionReportTemplateInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"Enter the template id for the template to be deleted.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteAttributionReportTemplateMCPTool creates the MCP Tool instance for DeleteAttributionReportTemplate
func NewDeleteAttributionReportTemplateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteAttributionReportTemplate",
		"Use this method to delete an existing template.\n\nPermissions required: Review Legal Obligations For Components Licenses for the root organization",
		[]byte(DeleteAttributionReportTemplateInputSchema),
	)
}

// DeleteAttributionReportTemplateHandler is the handler function for the DeleteAttributionReportTemplate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteAttributionReportTemplateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/licenseLegalMetadata/report-template/{id}", args, []string{"id"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DeleteAttributionReportTemplate"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
