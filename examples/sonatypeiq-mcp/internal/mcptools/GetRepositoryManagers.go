package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetRepositoryManagers tool
const GetRepositoryManagersInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetRepositoryManagers tool (Status: 200, Content-Type: application/json)
const GetRepositoryManagersResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains details of configured repository managers.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **repositoryManagers** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n      - **instanceId** (Type: string):\n      - **name** (Type: string):\n      - **productName** (Type: string):\n      - **productVersion** (Type: string):\n"

// NewGetRepositoryManagersMCPTool creates the MCP Tool instance for GetRepositoryManagers
func NewGetRepositoryManagersMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRepositoryManagers",
		"Use this method to retrieve all configured repository managers.\n\nPermissions required: View IQ Elements",
		[]byte(GetRepositoryManagersInputSchema),
	)
}

// GetRepositoryManagersHandler is the handler function for the GetRepositoryManagers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRepositoryManagersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/repositoryManagers", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRepositoryManagers")
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
