package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetTagsUsedByApplications tool
const GetTagsUsedByApplicationsInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetTagsUsedByApplications tool (Status: 200, Content-Type: application/json)
const GetTagsUsedByApplicationsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> A list of application categories or tags applied to applications. Each application category or tag consists of an id, name, description and color.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **organizationId** (Type: string):\n    - **color** (Type: string):\n    - **description** (Type: string):\n    - **id** (Type: string):\n    - **name** (Type: string):\n"

// NewGetTagsUsedByApplicationsMCPTool creates the MCP Tool instance for GetTagsUsedByApplications
func NewGetTagsUsedByApplicationsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetTagsUsedByApplications",
		"Grouping applications with similar characteristics into categories makes policy management easier. You can then create a policy that applies to a specific category. Use this method to retrieve a list of application categories.",
		[]byte(GetTagsUsedByApplicationsInputSchema),
	)
}

// GetTagsUsedByApplicationsHandler is the handler function for the GetTagsUsedByApplications tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetTagsUsedByApplicationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applicationCategories/application", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetTagsUsedByApplications")
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
