package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateAvatarFromTemporary3 tool
const CreateAvatarFromTemporary3InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"cropperOffsetX\": {\n          \"example\": 50,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"cropperOffsetY\": {\n          \"example\": 50,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"cropperWidth\": {\n          \"example\": 120,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"needsCropping\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"url\": {\n          \"example\": \"http://example.com/jira/secure/temporaryavatar?cropped=true\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"owningObjectId\": {\n      \"description\": \"Entity id where to change avatar\",\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"owningObjectId\",\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateAvatarFromTemporary3 tool (Status: 201, Content-Type: application/json)
const CreateAvatarFromTemporary3ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returns the created avatar.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **selected** (Type: boolean):\n  - **id** (Type: string):\n      - Example: '1000'\n  - **owner** (Type: string):\n      - Example: 'fred'\n"

// NewCreateAvatarFromTemporary3MCPTool creates the MCP Tool instance for CreateAvatarFromTemporary3
func NewCreateAvatarFromTemporary3MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateAvatarFromTemporary3",
		"Create avatar from temporary - Creates avatar from temporary",
		[]byte(CreateAvatarFromTemporary3InputSchema),
	)
}

// CreateAvatarFromTemporary3Handler is the handler function for the CreateAvatarFromTemporary3 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateAvatarFromTemporary3Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/universal_avatar/type/{type}/owner/{owningObjectId}/avatar", args, []string{"owningObjectId", "type"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateAvatarFromTemporary3")
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
