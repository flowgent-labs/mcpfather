package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetPropertyViaRestfulTable tool
const SetPropertyViaRestfulTableInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"a String containing the property key.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the SetPropertyViaRestfulTable tool (Status: 200, Content-Type: application/json)
const SetPropertyViaRestfulTableResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the property exists and the currently authenticated user has permission to edit it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **key** (Type: string):\n  - **value** (Type: string):\n  - **example** (Type: string):\n"

// NewSetPropertyViaRestfulTableMCPTool creates the MCP Tool instance for SetPropertyViaRestfulTable
func NewSetPropertyViaRestfulTableMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetPropertyViaRestfulTable",
		"Update an application property - Update an application property via PUT. The \"value\" field present in the PUT will override the existing value.",
		[]byte(SetPropertyViaRestfulTableInputSchema),
	)
}

// SetPropertyViaRestfulTableHandler is the handler function for the SetPropertyViaRestfulTable tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetPropertyViaRestfulTableHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/application-properties/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetPropertyViaRestfulTable"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
