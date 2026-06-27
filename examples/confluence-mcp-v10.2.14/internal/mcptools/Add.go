package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Add tool
const AddInputSchema = "{\n  \"properties\": {\n    \"labelName\": {\n      \"description\": \"the name of the label to be added (do not include any prefix, team: prefix assumed)\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"a string containing the key of the space\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"labelName\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Add tool (Status: 401, Content-Type: application/json)
const AddResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling user is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Add tool (Status: 403, Content-Type: application/json)
const AddResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> If the calling user does not have permission to add any label to the given space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Add tool (Status: 404, Content-Type: application/json)
const AddResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given spaceKey.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewAddMCPTool creates the MCP Tool instance for Add
func NewAddMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Add",
		"Add a category to a space - Adds a category the description of a given {@link Space} identified by spaceKey.\n\nExample request URI to add space category 'testCategory' to space with space key TEST:\n\n"+"\x60"+"https://example.com/confluence/rest/api/space/TEST/category/testCategory"+"\x60"+"",
		[]byte(AddInputSchema),
	)
}

// AddHandler is the handler function for the Add tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/space/{spaceKey}/category/{labelName}", args, []string{"labelName", "spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Add"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
