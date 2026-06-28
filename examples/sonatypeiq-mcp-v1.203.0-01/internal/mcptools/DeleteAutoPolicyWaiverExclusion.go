package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteAutoPolicyWaiverExclusion tool
const DeleteAutoPolicyWaiverExclusionInputSchema = "{\n  \"properties\": {\n    \"autoPolicyWaiverExclusionId\": {\n      \"description\": \"Enter the autoPolicyWaiverId to be deleted\",\n      \"type\": \"string\"\n    },\n    \"autoPolicyWaiverId\": {\n      \"description\": \"Enter the relevant Auto Policy Waiver ID.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. A waiver exclusion corresponding to the autoPolicyWaiverExclusionId provided and within the scope specified will be deleted.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"autoPolicyWaiverExclusionId\",\n    \"autoPolicyWaiverId\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteAutoPolicyWaiverExclusionMCPTool creates the MCP Tool instance for DeleteAutoPolicyWaiverExclusion
func NewDeleteAutoPolicyWaiverExclusionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteAutoPolicyWaiverExclusion",
		"Use this method to delete an auto policy waiver exclusion, specified by the autoPolicyWaiverExclusionId.\n\nPermissions required: Waive Policy Violations",
		[]byte(DeleteAutoPolicyWaiverExclusionInputSchema),
	)
}

// DeleteAutoPolicyWaiverExclusionHandler is the handler function for the DeleteAutoPolicyWaiverExclusion tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteAutoPolicyWaiverExclusionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/autoPolicyWaiverExclusions/{ownerType}/{ownerId}/{autoPolicyWaiverId}/{autoPolicyWaiverExclusionId}", args, []string{"autoPolicyWaiverExclusionId", "autoPolicyWaiverId", "ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteAutoPolicyWaiverExclusion")
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
