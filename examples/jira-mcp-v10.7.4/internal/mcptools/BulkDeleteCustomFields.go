package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the BulkDeleteCustomFields tool
const BulkDeleteCustomFieldsInputSchema = "{\n  \"properties\": {\n    \"ids\": {\n      \"description\": \"A list of custom field IDs to delete.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ids\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the BulkDeleteCustomFields tool (Status: 200, Content-Type: application/json)
const BulkDeleteCustomFieldsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if at least one custom field was deleted\n\n## Response Structure\n\n- Structure (Type: object):\n  - **deletedCustomFields** (Type: array):\n      - Example: '[\"customfield_10000\"]'\n    - **Items** (Type: string):\n        - Example: '[\"customfield_10000\"]'\n  - **message** (Type: string):\n      - Example: 'Custom fields bulk delete operation finished.'\n  - **notDeletedCustomFields** (Type: object):\n      - Example: '{\"customfield_10001\":\"Validation for custom field deletion failed.\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: '{\"customfield_10001\":\"Validation for custom field deletion failed.\"}'\n"

// NewBulkDeleteCustomFieldsMCPTool creates the MCP Tool instance for BulkDeleteCustomFields
func NewBulkDeleteCustomFieldsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"BulkDeleteCustomFields",
		"Delete custom fields in bulk - Deletes custom fields in bulk.",
		[]byte(BulkDeleteCustomFieldsInputSchema),
	)
}

// BulkDeleteCustomFieldsHandler is the handler function for the BulkDeleteCustomFields tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func BulkDeleteCustomFieldsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/customFields", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "BulkDeleteCustomFields"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
