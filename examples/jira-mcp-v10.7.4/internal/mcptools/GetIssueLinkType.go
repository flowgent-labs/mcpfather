package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueLinkType tool
const GetIssueLinkTypeInputSchema = "{\n  \"properties\": {\n    \"issueLinkTypeId\": {\n      \"description\": \"The issue link type id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueLinkTypeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueLinkType tool (Status: 200, Content-Type: application/json)
const GetIssueLinkTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the issue link type with the given id.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: string):\n      - Example: '10000'\n  - **inward** (Type: string):\n      - Example: 'is duplicated by'\n  - **name** (Type: string):\n      - Example: 'Duplicate'\n  - **outward** (Type: string):\n      - Example: 'duplicates'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issueLinkType/10000'\n"

// NewGetIssueLinkTypeMCPTool creates the MCP Tool instance for GetIssueLinkType
func NewGetIssueLinkTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueLinkType",
		"Get information about an issue link type - Returns for a given issue link type id all information about this issue link type.",
		[]byte(GetIssueLinkTypeInputSchema),
	)
}

// GetIssueLinkTypeHandler is the handler function for the GetIssueLinkType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueLinkTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issueLinkType/{issueLinkTypeId}", args, []string{"issueLinkTypeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueLinkType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
