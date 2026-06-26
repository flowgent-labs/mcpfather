package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Get tool
const GetInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the space properties. Default value: \\u003ccode\\u003eversion\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"key\": {\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"description\": \"the limit of the number of items to return, this may be restricted by fixed system limits.\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the space to find properties under. Required.\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"he start point of the collection to return.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"key\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Get tool (Status: 200, Content-Type: application/json)
const GetResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of the space property.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Get tool (Status: 404, Content-Type: application/json)
const GetResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given key, or no property with the given key, or if the calling user does not have permission to view the space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetMCPTool creates the MCP Tool instance for Get
func NewGetMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Get",
		"Get space property by key - Returns a space property. \n\nExample request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/api/space/TST/property/example-property-key?expand=space,version"+"\x60"+"",
		[]byte(GetInputSchema),
	)
}

// GetHandler is the handler function for the Get tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/property/{key}", args, []string{"key", "spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Get"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
