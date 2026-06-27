package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetData tool
const GetDataInputSchema = "{\n  \"properties\": {\n    \"applicationPublicId\": {\n      \"description\": \"Enter the applicationPublicId for the evaluated application.\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId (reportId) of the application report created after the evaluation. \",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationPublicId\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetDataMCPTool creates the MCP Tool instance for GetData
func NewGetDataMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetData",
		"This is an older version of the endpoint. This call will now be redirected to /api/v2/applications/{applicationPublicId}/reports/{scanId}/raw.",
		[]byte(GetDataInputSchema),
	)
}

// GetDataHandler is the handler function for the GetData tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetDataHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications/{applicationPublicId}/reports/{scanId}", args, []string{"applicationPublicId", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetData")
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
