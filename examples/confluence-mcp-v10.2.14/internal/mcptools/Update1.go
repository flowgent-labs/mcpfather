package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Update1 tool
const Update1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"the content property being updated\"\n    },\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the content properties. Default value: \\u003ccode\\u003eversion\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"description\": \"the key of the content property. Required.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"key\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Update1 tool (Status: 200, Content-Type: application/json)
const Update1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns  a JSON representation of the content property, or a 404 NOT FOUND if there is no content with the given id, or no property with the given key, or if the user is not permitted.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update1 tool (Status: 400, Content-Type: application/json)
const Update1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if the given property has a different ContentId to the one in the path, or if the property has a different key to the one in the path, or the value is missing, or the value is too long.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update1 tool (Status: 403, Content-Type: application/json)
const Update1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the user does not have permission to edit the content with the given ContentId.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update1 tool (Status: 404, Content-Type: application/json)
const Update1ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> eturned if there is no content with the given id, or no property with the given key, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update1 tool (Status: 409, Content-Type: application/json)
const Update1ResponseTemplate_E = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Returned if the given version is does not match the expected target version of the updated property.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update1 tool (Status: 413, Content-Type: application/json)
const Update1ResponseTemplate_F = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 413\n\n**Content-Type:** application/json\n\n> Returned if the value is too long.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdate1MCPTool creates the MCP Tool instance for Update1
func NewUpdate1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Update1",
		"Update content property - Updates a content property. The body contains the representation of the content property. Must include the property id, and the new version number. Attempts to create a new content property if the given version number is "+"\x60"+"1"+"\x60"+"",
		[]byte(Update1InputSchema),
	)
}

// Update1Handler is the handler function for the Update1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Update1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/content/{id}/property/{key}", args, []string{"id", "key"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Update1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
