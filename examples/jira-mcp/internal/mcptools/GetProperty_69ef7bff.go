package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetProperty_69ef7bff tool
const GetProperty_69ef7bffInputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"a String containing the property key.\",\n      \"type\": \"string\"\n    },\n    \"keyFilter\": {\n      \"description\": \"when fetching a list allows the list to be filtered by the property's start of key\\ne.g. \\\"jira.lf.*\\\" whould fetch only those permissions that are editable and whose keys start with\\n     *                        \\\"jira.lf.\\\". This is a regex.\",\n      \"type\": \"string\"\n    },\n    \"permissionLevel\": {\n      \"description\": \"when fetching a list specifies the permission level of all items in the list\\nsee {@link com.atlassian.jira.bc.admin.ApplicationPropertiesService.EditPermissionLevel}\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"permissionLevel\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProperty_69ef7bff tool (Status: 200, Content-Type: application/json)
const GetProperty_69ef7bffResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the property exists and the currently authenticated user has permission to view it. Contains a full representation of the property.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **key** (Type: string):\n  - **value** (Type: string):\n  - **example** (Type: string):\n"

// NewGetProperty_69ef7bffMCPTool creates the MCP Tool instance for GetProperty_69ef7bff
func NewGetProperty_69ef7bffMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProperty_69ef7bff",
		"Get an application property by key - Returns an application property.",
		[]byte(GetProperty_69ef7bffInputSchema),
	)
}

// GetProperty_69ef7bffHandler is the handler function for the GetProperty_69ef7bff tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProperty_69ef7bffHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/properties/{propertyKey}", args, []string{"boardId", "propertyKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetProperty_69ef7bff")
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
