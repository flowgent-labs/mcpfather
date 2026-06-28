package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeletePolicyWaiver tool
const DeletePolicyWaiverInputSchema = "{\n  \"properties\": {\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. A waiver corresponding to the policyWaiverId provided and within the scope specified will be deleted.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    },\n    \"policyWaiverId\": {\n      \"description\": \"Enter the policyWaiverId to be deleted.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\",\n    \"policyWaiverId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeletePolicyWaiverMCPTool creates the MCP Tool instance for DeletePolicyWaiver
func NewDeletePolicyWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeletePolicyWaiver",
		"Use this method to delete a waiver, specified by the policyWaiverId.\n\nPermissions required: Waive Policy Violations",
		[]byte(DeletePolicyWaiverInputSchema),
	)
}

// DeletePolicyWaiverHandler is the handler function for the DeletePolicyWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeletePolicyWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/policyWaivers/{ownerType}/{ownerId}/{policyWaiverId}", args, []string{"ownerId", "ownerType", "policyWaiverId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeletePolicyWaiver")
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
