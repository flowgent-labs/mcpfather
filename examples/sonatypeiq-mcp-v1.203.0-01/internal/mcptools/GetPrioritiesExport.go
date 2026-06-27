package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetPrioritiesExport tool
const GetPrioritiesExportInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the applicationId.\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetPrioritiesExportMCPTool creates the MCP Tool instance for GetPrioritiesExport
func NewGetPrioritiesExportMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPrioritiesExport",
		"Use this method to retrieve the priorities, by providing the applicationId and scanId.\n\nPermissions required: View IQ Elements",
		[]byte(GetPrioritiesExportInputSchema),
	)
}

// GetPrioritiesExportHandler is the handler function for the GetPrioritiesExport tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPrioritiesExportHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/developer/priorities/{applicationId}/{scanId}/export", args, []string{"applicationId", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPrioritiesExport")
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
