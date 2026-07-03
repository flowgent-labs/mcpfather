package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the UpdateOwnerArtifactoryConnectionStatus tool
const UpdateOwnerArtifactoryConnectionStatusInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Set values for the connection properties " + "\x60" + "enabled" + "\x60" + " and " + "\x60" + "allowOverride" + "\x60" + ".\",\n      \"properties\": {\n        \"allowOverride\": {\n          \"type\": \"boolean\"\n        },\n        \"enabled\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Enter the internal ID of the owner.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the owner type.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateOwnerArtifactoryConnectionStatusMCPTool creates the MCP Tool instance for UpdateOwnerArtifactoryConnectionStatus
func NewUpdateOwnerArtifactoryConnectionStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateOwnerArtifactoryConnectionStatus",
		"Use this method to enable/disable an existing Artifactory connection for the specified owner.\n\nPermissions required: Edit IQ Elements",
		[]byte(UpdateOwnerArtifactoryConnectionStatusInputSchema),
	)
}

// UpdateOwnerArtifactoryConnectionStatusHandler is the handler function for the UpdateOwnerArtifactoryConnectionStatus tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateOwnerArtifactoryConnectionStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/artifactoryConnection/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateOwnerArtifactoryConnectionStatus")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
