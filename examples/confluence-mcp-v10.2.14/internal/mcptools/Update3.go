package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Update3 tool
const Update3InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"space property to be updated\"\n    },\n    \"key\": {\n      \"description\": \"the key of the property\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the space to find properties under. Required.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Update3 tool (Status: 200, Content-Type: application/json)
const Update3ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of the property.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update3 tool (Status: 400, Content-Type: application/json)
const Update3ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if the given property has a different spaceKey to the one in the path, or if the property has a different key to the one in the path, or no property value was provided, or the value is too long.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update3 tool (Status: 403, Content-Type: application/json)
const Update3ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> returned if the user does not have permission to edit the space with the given spaceKey.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update3 tool (Status: 404, Content-Type: application/json)
const Update3ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given key, or no property with the given key, or if the calling user does not have permission to view the space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update3 tool (Status: 409, Content-Type: application/json)
const Update3ResponseTemplate_E = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Returned if the given version is does not match the expected target version of the updated property.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update3 tool (Status: 413, Content-Type: application/json)
const Update3ResponseTemplate_F = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 413\n\n**Content-Type:** application/json\n\n> Returned if the value is too long.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdate3MCPTool creates the MCP Tool instance for Update3
func NewUpdate3MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Update3",
		"Update space property - Updates a space property.The body contains the representation of the space property. Must include new version number.If the given version number is 1, attempts to create a new space property.",
		[]byte(Update3InputSchema),
	)
}

// Update3Handler is the handler function for the Update3 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Update3Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/space/{spaceKey}/property/{key}", args, []string{"key", "spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Update3"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
