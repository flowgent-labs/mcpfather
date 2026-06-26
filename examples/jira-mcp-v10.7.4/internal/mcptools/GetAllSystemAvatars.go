package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAllSystemAvatars tool
const GetAllSystemAvatarsInputSchema = "{\n  \"properties\": {\n    \"type\": {\n      \"description\": \"the avatar type\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAllSystemAvatars tool (Status: 200, Content-Type: application/json)
const GetAllSystemAvatarsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a map containing a list of system avatars. A map is returned to be consistent with the shape of the project/KEY/avatars REST end point.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **owner** (Type: string):\n      - Example: 'fred'\n  - **selected** (Type: boolean):\n  - **id** (Type: string):\n      - Example: '1000'\n"

// NewGetAllSystemAvatarsMCPTool creates the MCP Tool instance for GetAllSystemAvatars
func NewGetAllSystemAvatarsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllSystemAvatars",
		"Get all system avatars - Returns all system avatars of the given type.",
		[]byte(GetAllSystemAvatarsInputSchema),
	)
}

// GetAllSystemAvatarsHandler is the handler function for the GetAllSystemAvatars tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllSystemAvatarsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/avatar/{type}/system", args, []string{"type"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAllSystemAvatars"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
