package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteContainerImagePolicyWaiver tool
const DeleteContainerImagePolicyWaiverInputSchema = "{\n  \"properties\": {\n    \"containerImageId\": {\n      \"description\": \"Enter the container id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"containerImageId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteContainerImagePolicyWaiverMCPTool creates the MCP Tool instance for DeleteContainerImagePolicyWaiver
func NewDeleteContainerImagePolicyWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteContainerImagePolicyWaiver",
		"Use this method to delete a container waiver, specified by the containerImageId.\n\nPermissions required: Waive Policy Violations",
		[]byte(DeleteContainerImagePolicyWaiverInputSchema),
	)
}

// DeleteContainerImagePolicyWaiverHandler is the handler function for the DeleteContainerImagePolicyWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteContainerImagePolicyWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/firewall/container-image/{containerImageId}/policyWaiver", args, []string{"containerImageId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteContainerImagePolicyWaiver")
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
