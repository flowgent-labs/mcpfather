package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the ExportComponentSearchReport tool
const ExportComponentSearchReportInputSchema = "{\n  \"properties\": {\n    \"cveId\": {\n      \"description\": \"CVE identifier(s). Can be specified multiple times for multiple CVEs. Defaults to CVE-2025-55182 if not specified.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"type\": \"object\"\n}"

// NewExportComponentSearchReportMCPTool creates the MCP Tool instance for ExportComponentSearchReport
func NewExportComponentSearchReportMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ExportComponentSearchReport",
		"Export component search results as CSV (streaming). Identifies applications containing components affected by one or more CVEs. Multiple CVE IDs can be specified using multiple cveId query parameters (e.g., ?cveId=CVE-2025-1&cveId=CVE-2025-2). If no CVE ID is specified, defaults to CVE-2025-55182 (React2Shell) for backwards compatibility. Results are streamed to avoid memory issues with large datasets. Keep-alive mechanism prevents ALB timeouts during long-running queries. <p>Permissions Required: View IQ Elements",
		[]byte(ExportComponentSearchReportInputSchema),
	)
}

// ExportComponentSearchReportHandler is the handler function for the ExportComponentSearchReport tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ExportComponentSearchReportHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/componentSearch/downloadComponentSearchReport", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ExportComponentSearchReport")
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
