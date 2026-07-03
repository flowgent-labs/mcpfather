package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteAutoPolicyWaiver tool
const DeleteAutoPolicyWaiverInputSchema = "{\n  \"properties\": {\n    \"autoPolicyWaiverId\": {\n      \"description\": \"Enter the autoPolicyWaiverId to be deleted\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. A waiver corresponding to the autoPolicyWaiverId provided and within the scope specified will be deleted.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"autoPolicyWaiverId\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteAutoPolicyWaiverMCPTool creates the MCP Tool instance for DeleteAutoPolicyWaiver
func NewDeleteAutoPolicyWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteAutoPolicyWaiver",
		"Use this method to delete an auto policy waiver, specified by the autoPolicyWaiverId.\n\nPermissions required: Waive Policy Violations",
		[]byte(DeleteAutoPolicyWaiverInputSchema),
	)
}

// DeleteAutoPolicyWaiverHandler is the handler function for the DeleteAutoPolicyWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteAutoPolicyWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/autoPolicyWaivers/{ownerType}/{ownerId}/{autoPolicyWaiverId}", args, []string{"autoPolicyWaiverId", "ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteAutoPolicyWaiver")
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
