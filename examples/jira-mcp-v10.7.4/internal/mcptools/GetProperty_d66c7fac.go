package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProperty_d66c7fac tool
const GetProperty_d66c7facInputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"a String containing the property key.\",\n      \"type\": \"string\"\n    },\n    \"keyFilter\": {\n      \"description\": \"when fetching a list allows the list to be filtered by the property's start of key\\ne.g. \\\"jira.lf.*\\\" whould fetch only those permissions that are editable and whose keys start with\\n     *                        \\\"jira.lf.\\\". This is a regex.\",\n      \"type\": \"string\"\n    },\n    \"permissionLevel\": {\n      \"description\": \"when fetching a list specifies the permission level of all items in the list\\nsee {@link com.atlassian.jira.bc.admin.ApplicationPropertiesService.EditPermissionLevel}\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"permissionLevel\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProperty_d66c7fac tool (Status: 200, Content-Type: application/json)
const GetProperty_d66c7facResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the property exists and the currently authenticated user has permission to view it. Contains a full representation of the property.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **example** (Type: string):\n  - **key** (Type: string):\n  - **value** (Type: string):\n"

// NewGetProperty_d66c7facMCPTool creates the MCP Tool instance for GetProperty_d66c7fac
func NewGetProperty_d66c7facMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProperty_d66c7fac",
		"Get an application property by key - Returns an application property.",
		[]byte(GetProperty_d66c7facInputSchema),
	)
}

// GetProperty_d66c7facHandler is the handler function for the GetProperty_d66c7fac tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProperty_d66c7facHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/application-properties", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProperty_d66c7fac"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
