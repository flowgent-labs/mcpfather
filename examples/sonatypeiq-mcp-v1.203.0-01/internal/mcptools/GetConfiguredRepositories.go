package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguredRepositories tool
const GetConfiguredRepositoriesInputSchema = "{\n  \"properties\": {\n    \"repositoryManagerId\": {\n      \"description\": \"Enter the repository manager ID.\",\n      \"type\": \"string\"\n    },\n    \"sinceUtcTimestamp\": {\n      \"description\": \"Enter the epoch time in milliseconds when the repository was last updated.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"repositoryManagerId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguredRepositories tool (Status: 200, Content-Type: application/json)
const GetConfiguredRepositoriesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the configuration details of the requested repository manager.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **repositories** (Type: array):\n    - **Items** (Type: object):\n      - **auditEnabled** (Type: boolean):\n      - **format** (Type: string):\n      - **namespaceConfusionProtectionEnabled** (Type: boolean):\n      - **policyCompliantComponentSelectionEnabled** (Type: boolean):\n      - **publicId** (Type: string):\n      - **quarantineEnabled** (Type: boolean):\n      - **repositoryId** (Type: string):\n      - **type** (Type: string):\n"

// NewGetConfiguredRepositoriesMCPTool creates the MCP Tool instance for GetConfiguredRepositories
func NewGetConfiguredRepositoriesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguredRepositories",
		"Use this method to retrieve the configuration details of an existing repository manager.\n\nPermissions required: View IQ Elements",
		[]byte(GetConfiguredRepositoriesInputSchema),
	)
}

// GetConfiguredRepositoriesHandler is the handler function for the GetConfiguredRepositories tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguredRepositoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/repositories/configuration/{repositoryManagerId}", args, []string{"repositoryManagerId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguredRepositories")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
