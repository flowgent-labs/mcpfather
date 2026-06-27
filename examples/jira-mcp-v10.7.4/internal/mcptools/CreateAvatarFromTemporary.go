package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the CreateAvatarFromTemporary tool
const CreateAvatarFromTemporaryInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"cropping instructions\",\n      \"properties\": {\n        \"cropperOffsetX\": {\n          \"example\": 50,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"cropperOffsetY\": {\n          \"example\": 50,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"cropperWidth\": {\n          \"example\": 120,\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"needsCropping\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"url\": {\n          \"example\": \"http://example.com/jira/secure/temporaryavatar?cropped=true\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"type\": {\n      \"description\": \"the avatar type\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// NewCreateAvatarFromTemporaryMCPTool creates the MCP Tool instance for CreateAvatarFromTemporary
func NewCreateAvatarFromTemporaryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateAvatarFromTemporary",
		"Update avatar cropping - Updates the cropping instructions of the temporary avatar",
		[]byte(CreateAvatarFromTemporaryInputSchema),
	)
}

// CreateAvatarFromTemporaryHandler is the handler function for the CreateAvatarFromTemporary tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateAvatarFromTemporaryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/avatar/{type}/temporaryCrop", args, []string{"type"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateAvatarFromTemporary"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
