package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteComponentLabel tool
const DeleteComponentLabelInputSchema = "{\n  \"properties\": {\n    \"componentHash\": {\n      \"description\": \"Enter the SHA1 hash of the component.\",\n      \"type\": \"string\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Possible values : applicationId or organizationId\",\n      \"type\": \"string\"\n    },\n    \"labelName\": {\n      \"description\": \"Enter the label name to un-assign from this component.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Possible values: application or organization\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"componentHash\",\n    \"internalOwnerId\",\n    \"labelName\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteComponentLabelMCPTool creates the MCP Tool instance for DeleteComponentLabel
func NewDeleteComponentLabelMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteComponentLabel",
		"Use this method to un-assign a label from a component.",
		[]byte(DeleteComponentLabelInputSchema),
	)
}

// DeleteComponentLabelHandler is the handler function for the DeleteComponentLabel tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteComponentLabelHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/components/{componentHash}/labels/{labelName}/{ownerType}s/{internalOwnerId}", args, []string{"componentHash", "internalOwnerId", "labelName", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteComponentLabel")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
