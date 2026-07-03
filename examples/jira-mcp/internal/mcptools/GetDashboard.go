package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetDashboard tool
const GetDashboardInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The dashboard id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetDashboard tool (Status: 200, Content-Type: application/json)
const GetDashboardResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a single dashboard.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/dashboard/10000'\n  - **view** (Type: string):\n      - Example: 'http://localhost:8090/jira/secure/Dashboard.jspa?selectPageId=10000'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **name** (Type: string):\n      - Example: 'System Dashboard'\n"

// NewGetDashboardMCPTool creates the MCP Tool instance for GetDashboard
func NewGetDashboardMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetDashboard",
		"Get a single dashboard by ID - Returns a single dashboard.",
		[]byte(GetDashboardInputSchema),
	)
}

// GetDashboardHandler is the handler function for the GetDashboard tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetDashboardHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/dashboard/{id}", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetDashboard")
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
