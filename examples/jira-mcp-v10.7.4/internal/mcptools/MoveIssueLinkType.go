package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the MoveIssueLinkType tool
const MoveIssueLinkTypeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The new position to move the issue link type\",\n      \"properties\": {\n        \"newPosition\": {\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueLinkTypeId\": {\n      \"description\": \"Id of the issue link type to move.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"issueLinkTypeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the MoveIssueLinkType tool (Status: 200, Content-Type: application/json)
const MoveIssueLinkTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the updated issue link type.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'Duplicate'\n  - **outward** (Type: string):\n      - Example: 'duplicates'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issueLinkType/10000'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **inward** (Type: string):\n      - Example: 'is duplicated by'\n"

// NewMoveIssueLinkTypeMCPTool creates the MCP Tool instance for MoveIssueLinkType
func NewMoveIssueLinkTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"MoveIssueLinkType",
		"Update the order of the issue link type. - Moves the issue link type to a new position within the list.",
		[]byte(MoveIssueLinkTypeInputSchema),
	)
}

// MoveIssueLinkTypeHandler is the handler function for the MoveIssueLinkType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveIssueLinkTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issueLinkType/{issueLinkTypeId}/order", args, []string{"issueLinkTypeId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "MoveIssueLinkType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
