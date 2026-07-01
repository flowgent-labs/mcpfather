package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the ExpandForHumans tool
const ExpandForHumansInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"the id of the attachment to expand.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ExpandForHumans tool (Status: 200, Content-Type: application/json)
const ExpandForHumansResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> JSON representation of the attachment expanded contents. Empty entry list means that attachment cannot be expanded. It's either empty, corrupt or not an archive at all.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **entries** (Type: object):\n  - **id** (Type: integer, int64):\n      - Example: '7237823'\n  - **mediaType** (Type: string):\n      - Example: 'application/zip'\n  - **name** (Type: string):\n      - Example: 'images.zip'\n  - **totalEntryCount** (Type: integer, int64):\n      - Example: '39'\n"

// NewExpandForHumansMCPTool creates the MCP Tool instance for ExpandForHumans
func NewExpandForHumansMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ExpandForHumans",
		"Get human-readable attachment expansion - Tries to expand an attachment. Output is human-readable and subject to change.",
		[]byte(ExpandForHumansInputSchema),
	)
}

// ExpandForHumansHandler is the handler function for the ExpandForHumans tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ExpandForHumansHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/attachment/{id}/expand/human", args, []string{"id"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ExpandForHumans")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
