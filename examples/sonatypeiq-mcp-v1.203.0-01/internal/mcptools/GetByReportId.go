package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetByReportId tool
const GetByReportIdInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal applicationId for the application you want to generate the SBOM. You can also retrieve the applicationId using the Application REST API.\",\n      \"type\": \"string\"\n    },\n    \"cdxVersion\": {\n      \"description\": \"Possible values are 1.1|1.2|1.3|1.4|1.5|1.6.\",\n      \"pattern\": \"1.1|1.2|1.3|1.4|1.5|1.6\",\n      \"type\": \"string\"\n    },\n    \"reportId\": {\n      \"description\": \"Enter the reportId to generate the SBOM for the application for a specific scan report.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"cdxVersion\",\n    \"reportId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetByReportIdMCPTool creates the MCP Tool instance for GetByReportId
func NewGetByReportIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetByReportId",
		"Use this method to generate a CycloneDX SBOM for an application.<p>Permissions Required: View IQ Elements",
		[]byte(GetByReportIdInputSchema),
	)
}

// GetByReportIdHandler is the handler function for the GetByReportId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetByReportIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/cycloneDx/{cdxVersion}/{applicationId}/reports/{reportId}", args, []string{"applicationId", "cdxVersion", "reportId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetByReportId")
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
