package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the CreateAvatarFromTemporary2 tool
const CreateAvatarFromTemporary2InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Cropping instructions\",\n      \"properties\": {\n        \"cropperOffsetX\": {\n          \"example\": 50,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"cropperOffsetY\": {\n          \"example\": 50,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"cropperWidth\": {\n          \"example\": 120,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"needsCropping\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"url\": {\n          \"example\": \"http://example.com/jira/secure/temporaryavatar?cropped=true\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"Project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateAvatarFromTemporary2 tool (Status: 201, Content-Type: application/json)
const CreateAvatarFromTemporary2ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Avatar data\n\n## Response Structure\n\n- Structure (Type: object):\n  - **owner** (Type: string):\n      - Example: 'fred'\n  - **selected** (Type: boolean):\n  - **id** (Type: string):\n      - Example: '1000'\n"

// NewCreateAvatarFromTemporary2MCPTool creates the MCP Tool instance for CreateAvatarFromTemporary2
func NewCreateAvatarFromTemporary2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateAvatarFromTemporary2",
		"Create avatar from temporary - Converts the temporary avatar into the final one. This is step 2/3 of changing an avatar for a project:\n- Upload (store temporary avatar)\n- Crop (create avatar from temporary)\n- Update (update project avatar)",
		[]byte(CreateAvatarFromTemporary2InputSchema),
	)
}

// CreateAvatarFromTemporary2Handler is the handler function for the CreateAvatarFromTemporary2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateAvatarFromTemporary2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/project/{projectIdOrKey}/avatar", args, []string{"projectIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateAvatarFromTemporary2"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
