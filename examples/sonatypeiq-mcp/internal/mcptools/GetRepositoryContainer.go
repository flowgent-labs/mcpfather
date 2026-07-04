package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetRepositoryContainer tool
const GetRepositoryContainerInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetRepositoryContainer tool (Status: 200, Content-Type: application/json)
const GetRepositoryContainerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the ID and name for the repository container.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **relatedOrganizationId** (Type: string):\n  - **id** (Type: string):\n  - **name** (Type: string):\n"

// NewGetRepositoryContainerMCPTool creates the MCP Tool instance for GetRepositoryContainer
func NewGetRepositoryContainerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRepositoryContainer",
		"Use this method to retrieve the ID and name for the repository container.\n\nPermissions required: View IQ Elements",
		[]byte(GetRepositoryContainerInputSchema),
	)
}

// GetRepositoryContainerHandler is the handler function for the GetRepositoryContainer tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRepositoryContainerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/repositoryContainer", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRepositoryContainer")
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
