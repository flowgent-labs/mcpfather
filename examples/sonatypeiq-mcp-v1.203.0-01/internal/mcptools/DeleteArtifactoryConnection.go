package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteArtifactoryConnection tool
const DeleteArtifactoryConnectionInputSchema = "{\n  \"properties\": {\n    \"artifactoryConnectionId\": {\n      \"description\": \"Enter the Artifactory connection ID.\",\n      \"type\": \"string\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Enter the internal ID of the owner.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the owner type.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"artifactoryConnectionId\",\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteArtifactoryConnectionMCPTool creates the MCP Tool instance for DeleteArtifactoryConnection
func NewDeleteArtifactoryConnectionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteArtifactoryConnection",
		"Use this method to delete an existing Artifactory connection.\n\nPermissions required: Edit IQ Elements",
		[]byte(DeleteArtifactoryConnectionInputSchema),
	)
}

// DeleteArtifactoryConnectionHandler is the handler function for the DeleteArtifactoryConnection tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteArtifactoryConnectionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/config/artifactoryConnection/{ownerType}/{internalOwnerId}/{artifactoryConnectionId}", args, []string{"artifactoryConnectionId", "internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteArtifactoryConnection")
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
