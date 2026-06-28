package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the ConfigureRepositories tool
const ConfigureRepositoriesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Enter values for the repository configuration properties to be updated.\",\n      \"properties\": {\n        \"repositories\": {\n          \"items\": {\n            \"properties\": {\n              \"auditEnabled\": {\n                \"type\": \"boolean\"\n              },\n              \"format\": {\n                \"type\": \"string\"\n              },\n              \"namespaceConfusionProtectionEnabled\": {\n                \"type\": \"boolean\"\n              },\n              \"policyCompliantComponentSelectionEnabled\": {\n                \"type\": \"boolean\"\n              },\n              \"publicId\": {\n                \"type\": \"string\"\n              },\n              \"quarantineEnabled\": {\n                \"type\": \"boolean\"\n              },\n              \"repositoryId\": {\n                \"type\": \"string\"\n              },\n              \"type\": {\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"repositoryManagerId\": {\n      \"description\": \"Enter the repository manager ID.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"repositoryManagerId\"\n  ],\n  \"type\": \"object\"\n}"

// NewConfigureRepositoriesMCPTool creates the MCP Tool instance for ConfigureRepositories
func NewConfigureRepositoriesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ConfigureRepositories",
		"Use this method to update the repositories for an existing repository manager.\n\nPermissions required: Edit IQ Elements",
		[]byte(ConfigureRepositoriesInputSchema),
	)
}

// ConfigureRepositoriesHandler is the handler function for the ConfigureRepositories tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ConfigureRepositoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/firewall/repositories/configuration/{repositoryManagerId}", args, []string{"repositoryManagerId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ConfigureRepositories")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
