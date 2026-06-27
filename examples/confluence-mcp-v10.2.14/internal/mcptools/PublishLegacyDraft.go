package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the PublishLegacyDraft tool
const PublishLegacyDraftInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The content to be created, where the status of the included content is \\\"current\\\",\\nand the content has an ID (which will be the draft ID)\"\n    },\n    \"draftId\": {\n      \"description\": \"the id of the draft\",\n      \"type\": \"string\"\n    },\n    \"expand\": {\n      \"description\": \"A comma separated list of properties to expand on the content. Default value: \\u003ccode\\u003ebody.storage,history,space,version,ancestors\\u003c/code\\u003e\",\n      \"type\": \"string\"\n    },\n    \"status\": {\n      \"description\": \"only support 'draft' status for now.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"draftId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PublishLegacyDraft tool (Status: 200, Content-Type: application/json)
const PublishLegacyDraftResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a JSON representation of the content\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewPublishLegacyDraftMCPTool creates the MCP Tool instance for PublishLegacyDraft
func NewPublishLegacyDraftMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PublishLegacyDraft",
		"Publish legacy draft - Publishes a legacy draft of a Content created from a ContentBlueprint.",
		[]byte(PublishLegacyDraftInputSchema),
	)
}

// PublishLegacyDraftHandler is the handler function for the PublishLegacyDraft tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PublishLegacyDraftHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/content/blueprint/instance/{draftId}", args, []string{"draftId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "PublishLegacyDraft"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
