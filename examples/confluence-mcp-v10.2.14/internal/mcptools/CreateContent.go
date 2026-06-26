package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the CreateContent tool
const CreateContentInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"new content to be created.\"\n    },\n    \"expand\": {\n      \"description\": \" comma separated list of properties to expand on the content. Default value: \\u003ccode\\u003ehistory,space,version\\u003c/code\\u003e\",\n      \"type\": \"string\"\n    },\n    \"status\": {\n      \"description\": \"list of Content statuses to filter results on. \\n\\n Default value: \\u003ccode\\u003e[current]\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateContent tool (Status: 200, Content-Type: application/json)
const CreateContentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a JSON representation of the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateContent tool (Status: 404, Content-Type: application/json)
const CreateContentResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> returned if there is no content with the given id or if the user is not permitted.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreateContentMCPTool creates the MCP Tool instance for CreateContent
func NewCreateContentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateContent",
		"Create content - Creates a new piece of Content or publishes the draft if the content id is present. For the case publishing draft, a new piece of content will be created and all metadata from the draft will be transferred into the newly created content.",
		[]byte(CreateContentInputSchema),
	)
}

// CreateContentHandler is the handler function for the CreateContent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateContentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/content", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateContent"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
