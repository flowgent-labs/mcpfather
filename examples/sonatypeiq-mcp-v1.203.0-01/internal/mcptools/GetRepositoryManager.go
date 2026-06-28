package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetRepositoryManager tool
const GetRepositoryManagerInputSchema = "{\n  \"properties\": {\n    \"repositoryManagerId\": {\n      \"description\": \"Enter the repository manager ID.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"repositoryManagerId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRepositoryManager tool (Status: 200, Content-Type: application/json)
const GetRepositoryManagerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of the repository manager requested.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **instanceId** (Type: string):\n  - **name** (Type: string):\n  - **productName** (Type: string):\n  - **productVersion** (Type: string):\n  - **id** (Type: string):\n"

// NewGetRepositoryManagerMCPTool creates the MCP Tool instance for GetRepositoryManager
func NewGetRepositoryManagerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRepositoryManager",
		"Use this method to retrieve details of an existing repository manager.\n\nPermissions required: View IQ Elements",
		[]byte(GetRepositoryManagerInputSchema),
	)
}

// GetRepositoryManagerHandler is the handler function for the GetRepositoryManager tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRepositoryManagerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/repositoryManagers/{repositoryManagerId}", args, []string{"repositoryManagerId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRepositoryManager")
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
