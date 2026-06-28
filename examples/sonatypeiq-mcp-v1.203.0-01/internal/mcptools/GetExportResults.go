package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetExportResults tool
const GetExportResultsInputSchema = "{\n  \"properties\": {\n    \"allComponents\": {\n      \"default\": false,\n      \"description\": \"Set to " + "\x60" + "true" + "\x60" + " to retrieve results that include components with no violations.\",\n      \"type\": \"boolean\"\n    },\n    \"mode\": {\n      \"enum\": [\n        \"sbomManager\"\n      ],\n      \"type\": \"string\"\n    },\n    \"page\": {\n      \"description\": \"Enter the page no. for the page containing results\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"description\": \"Enter the no. of results that should be visible per page, unset gives all results\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"query\": {\n      \"description\": \"A well-formed search query.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"query\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetExportResultsMCPTool creates the MCP Tool instance for GetExportResults
func NewGetExportResultsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetExportResults",
		"Use this method to generate a csv file containing your search results. The default delimiter in the generated file is comma. Use the advancedSearchCSVExportDelimiter property of the Configuration REST API to change the delimiter in the generated file.",
		[]byte(GetExportResultsInputSchema),
	)
}

// GetExportResultsHandler is the handler function for the GetExportResults tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetExportResultsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/search/advanced/export/csv", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetExportResults")
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
