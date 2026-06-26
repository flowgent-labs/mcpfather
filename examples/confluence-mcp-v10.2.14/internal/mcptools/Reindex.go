package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Reindex tool
const ReindexInputSchema = "{\n  \"properties\": {\n    \"option\": {\n      \"description\": \"The reindex options to control what content types are indexed.\\nAvailable options:\\n- CONTENT_ONLY: Index only content (pages, blog posts, etc.)\\n- ATTACHMENT_ONLY: Index only file attachments\\n- USER_ONLY: Index only user information (Only relevant for whole site reindexing, not for specific spaces)\\n\\nIf no options are specified, a full reindex of all relevant content types will be performed.\\nMultiple options can be specified to index specific combinations of content types.\\n\",\n      \"items\": {\n        \"enum\": [\n          \"CONTENT_ONLY\",\n          \"ATTACHMENT_ONLY\",\n          \"USER_ONLY\"\n        ],\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"spaceKey\": {\n      \"description\": \"Optional space keys to limit the reindex to specific spaces.\\nIf specified, only content within these spaces will be reindexed.\\nIf no space key's are specified, the entire site will be reindexed.\\nMultiple space keys can be provided to reindex multiple spaces.\\nIf a space key does not match any existing space, it will be silently ignored.\\n\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Reindex tool (Status: 200, Content-Type: application/json)
const ReindexResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Message indicating that the reindex task was successfully queued.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Reindex tool (Status: 400, Content-Type: application/json)
const ReindexResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if the request parameters are invalid.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Reindex tool (Status: 409, Content-Type: application/json)
const ReindexResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Returned if a reindex operation is already in progress.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Reindex tool (Status: 503, Content-Type: application/json)
const ReindexResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 503\n\n**Content-Type:** application/json\n\n> Returned if the reindexing was interrupted\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewReindexMCPTool creates the MCP Tool instance for Reindex
func NewReindexMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Reindex",
		"Rebuild Confluence search index - Rebuilds Confluence's search index.\nThis operation is only available to system administrators and may take significant time to complete.\n\nExample request URI(s):\n- "+"\x60"+"http://example.com/confluence/rest/api/index/reindex"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/index/reindex?option=CONTENT_ONLY&spaceKey=DEMO"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/index/reindex?option=ATTACHMENT_ONLY&option=CONTENT_ONLY&spaceKey=DEMO&spaceKey=TEST"+"\x60"+"\n",
		[]byte(ReindexInputSchema),
	)
}

// ReindexHandler is the handler function for the Reindex tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ReindexHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/index/reindex", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Reindex"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
