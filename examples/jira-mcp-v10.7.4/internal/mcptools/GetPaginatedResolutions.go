package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPaginatedResolutions tool
const GetPaginatedResolutionsInputSchema = "{\n  \"properties\": {\n    \"maxResults\": {\n      \"default\": 100,\n      \"description\": \"The maximum number of statuses to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"query\": {\n      \"default\": \"\",\n      \"description\": \"The string that status names will be matched with.\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"default\": 0,\n      \"description\": \"The index of the first status to return.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPaginatedResolutions tool (Status: 200, Content-Type: application/json)
const GetPaginatedResolutionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns paginated list of resolutions.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewGetPaginatedResolutionsMCPTool creates the MCP Tool instance for GetPaginatedResolutions
func NewGetPaginatedResolutionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPaginatedResolutions",
		"Get paginated filtered resolutions - Returns paginated list of filtered resolutions.",
		[]byte(GetPaginatedResolutionsInputSchema),
	)
}

// GetPaginatedResolutionsHandler is the handler function for the GetPaginatedResolutions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPaginatedResolutionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/resolution/page", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPaginatedResolutions"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
