package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetAuditLogs tool
const GetAuditLogsInputSchema = "{\n  \"properties\": {\n    \"endUtcDate\": {\n      \"description\": \"Enter the end UTC date in the format (yyyy-mm-dd).\",\n      \"type\": \"string\"\n    },\n    \"startUtcDate\": {\n      \"description\": \"Enter the start UTC date in the format (yyyy-mm-dd).\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewGetAuditLogsMCPTool creates the MCP Tool instance for GetAuditLogs
func NewGetAuditLogsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAuditLogs",
		"Use this method to retrieve the audit events for the specified time period.\n\nPermissions required: Access Audit Log",
		[]byte(GetAuditLogsInputSchema),
	)
}

// GetAuditLogsHandler is the handler function for the GetAuditLogs tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAuditLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/auditLogs", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAuditLogs")
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
