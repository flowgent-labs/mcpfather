package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the StoreTemporaryAvatar tool
const StoreTemporaryAvatarInputSchema = "{\n  \"properties\": {\n    \"filename\": {\n      \"description\": \"name of file being uploaded\",\n      \"type\": \"string\"\n    },\n    \"size\": {\n      \"description\": \"size of file\",\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"description\": \"the avatar type\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the StoreTemporaryAvatar tool (Status: 201, Content-Type: application/json)
const StoreTemporaryAvatarResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> temporary avatar cropping instructions\n\n## Response Structure\n\n- Structure (Type: object):\n  - **cropperOffsetY** (Type: integer, int32):\n      - Example: '50'\n  - **cropperWidth** (Type: integer, int32):\n      - Example: '120'\n  - **needsCropping** (Type: boolean):\n      - Example: 'true'\n  - **url** (Type: string):\n      - Example: 'http://example.com/jira/secure/temporaryavatar?cropped=true'\n  - **cropperOffsetX** (Type: integer, int32):\n      - Example: '50'\n"

// NewStoreTemporaryAvatarMCPTool creates the MCP Tool instance for StoreTemporaryAvatar
func NewStoreTemporaryAvatarMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"StoreTemporaryAvatar",
		"Create temporary avatar - Creates temporary avatar",
		[]byte(StoreTemporaryAvatarInputSchema),
	)
}

// StoreTemporaryAvatarHandler is the handler function for the StoreTemporaryAvatar tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func StoreTemporaryAvatarHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/avatar/{type}/temporary", args, []string{"type"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "StoreTemporaryAvatar")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
