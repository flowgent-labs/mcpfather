package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the UpdateCpeMatchingConfiguration tool
const UpdateCpeMatchingConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"allowOverride\": {\n          \"type\": \"boolean\"\n        },\n        \"enabled\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"internalOwnerId\": {\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateCpeMatchingConfigurationMCPTool creates the MCP Tool instance for UpdateCpeMatchingConfiguration
func NewUpdateCpeMatchingConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateCpeMatchingConfiguration",
		"Use this method to apply a given cpe matching configuration to an organization or application.<p>Permissions Required: Edit IQ Elements",
		[]byte(UpdateCpeMatchingConfigurationInputSchema),
	)
}

// UpdateCpeMatchingConfigurationHandler is the handler function for the UpdateCpeMatchingConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateCpeMatchingConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/{ownerType}/{internalOwnerId}/configuration/publicSource/cpe", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateCpeMatchingConfiguration")
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
