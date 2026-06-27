package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Create4 tool
const Create4InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"space property to be created\"\n    },\n    \"key\": {\n      \"description\": \"property key of the property to be created\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Create4 tool (Status: 200, Content-Type: application/json)
const Create4ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of the space property.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Create4 tool (Status: 400, Content-Type: application/json)
const Create4ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if the space already has a value with the given key, or no property value was provided, or the value is too long.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Create4 tool (Status: 403, Content-Type: application/json)
const Create4ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the user does not have permission to edit the space with the given key.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Create4 tool (Status: 413, Content-Type: application/json)
const Create4ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 413\n\n**Content-Type:** application/json\n\n> Returned if the value is too long.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreate4MCPTool creates the MCP Tool instance for Create4
func NewCreate4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Create4",
		"Create a space property with a specific key - Create a space property with a specific key.",
		[]byte(Create4InputSchema),
	)
}

// Create4Handler is the handler function for the Create4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Create4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/space/{spaceKey}/property/{key}", args, []string{"key", "spaceKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Create4"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
