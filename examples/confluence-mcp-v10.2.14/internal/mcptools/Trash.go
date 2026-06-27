package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Trash tool
const TrashInputSchema = "{\n  \"properties\": {\n    \"spaceKey\": {\n      \"description\": \"the key of the space to update.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Trash tool (Status: 403, Content-Type: application/json)
const TrashResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if user does not have correct permission\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Trash tool (Status: 404, Content-Type: application/json)
const TrashResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given key\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewTrashMCPTool creates the MCP Tool instance for Trash
func NewTrashMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Trash",
		"Remove all trash contents - Remove all content from the trash in the given space, deleting them permanently.Example request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/api/space/TEST/trash"+"\x60"+"",
		[]byte(TrashInputSchema),
	)
}

// TrashHandler is the handler function for the Trash tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func TrashHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/confluence/rest/api/space/{spaceKey}/trash", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Trash"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
