package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAvatars tool
const GetAvatarsInputSchema = "{\n  \"properties\": {\n    \"owningObjectId\": {\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"owningObjectId\",\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAvatars tool (Status: 200, Content-Type: application/json)
const GetAvatarsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all Jira avatars in JSON format, that are visible to the user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **owner** (Type: string):\n      - Example: 'fred'\n  - **selected** (Type: boolean):\n  - **id** (Type: string):\n      - Example: '1000'\n"

// NewGetAvatarsMCPTool creates the MCP Tool instance for GetAvatars
func NewGetAvatarsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAvatars",
		"Get all avatars for a type and owner - Returns a list of all avatars",
		[]byte(GetAvatarsInputSchema),
	)
}

// GetAvatarsHandler is the handler function for the GetAvatars tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAvatarsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/universal_avatar/type/{type}/owner/{owningObjectId}", args, []string{"owningObjectId", "type"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAvatars"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
