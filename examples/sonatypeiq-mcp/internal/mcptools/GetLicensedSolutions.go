package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetLicensedSolutions tool
const GetLicensedSolutionsInputSchema = "{\n  \"properties\": {\n    \"allowRelativeUrls\": {\n      \"default\": false,\n      \"description\": \"Whether or not relative URLs should be allowed.\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetLicensedSolutions tool (Status: 200, Content-Type: application/json)
const GetLicensedSolutionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successfully retrieved the list of licensed solutions.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **id** (Type: string):\n    - **url** (Type: string):\n"

// NewGetLicensedSolutionsMCPTool creates the MCP Tool instance for GetLicensedSolutions
func NewGetLicensedSolutionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLicensedSolutions",
		"Retrieves a list of licensed solutions. The base URL must be set to get results unless relative URLs are allowed.\n\nPermissions required: None ",
		[]byte(GetLicensedSolutionsInputSchema),
	)
}

// GetLicensedSolutionsHandler is the handler function for the GetLicensedSolutions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLicensedSolutionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/solutions/licensed", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLicensedSolutions")
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
