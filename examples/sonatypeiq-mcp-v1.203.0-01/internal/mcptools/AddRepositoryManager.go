package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddRepositoryManager tool
const AddRepositoryManagerInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Enter values for the new repository manager.\",\n      \"properties\": {\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"instanceId\": {\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"productName\": {\n          \"type\": \"string\"\n        },\n        \"productVersion\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddRepositoryManager tool (Status: 200, Content-Type: application/json)
const AddRepositoryManagerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of the new repository manager.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: string):\n  - **instanceId** (Type: string):\n  - **name** (Type: string):\n  - **productName** (Type: string):\n  - **productVersion** (Type: string):\n"

// NewAddRepositoryManagerMCPTool creates the MCP Tool instance for AddRepositoryManager
func NewAddRepositoryManagerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddRepositoryManager",
		"Use this method to add a new repository manager.\n\nPermissions required: Edit IQ Elements",
		[]byte(AddRepositoryManagerInputSchema),
	)
}

// AddRepositoryManagerHandler is the handler function for the AddRepositoryManager tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddRepositoryManagerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/firewall/repositoryManagers", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddRepositoryManager")
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
