package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPrioritySchemes tool
const GetPrioritySchemesInputSchema = "{\n  \"properties\": {\n    \"maxResults\": {\n      \"description\": \"how many results on the page should be included. Defaults to 100, maximum is 1000.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"startAt\": {\n      \"description\": \"the page offset, if not specified then defaults to 0\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPrioritySchemes tool (Status: 200, Content-Type: application/json)
const GetPrioritySchemesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Priority schemes\n\n## Response Structure\n\n- Structure (Type: object):\n  - **schemes** (Type: array):\n    - **Items** (Type: object):\n      - **description** (Type: string):\n      - **id** (Type: integer, int64):\n      - **name** (Type: string):\n      - **optionIds** (Type: array):\n        - **Items** (Type: string):\n      - **projectKeys** (Type: array):\n        - **Items** (Type: string):\n      - **self** (Type: string, uri):\n      - **defaultOptionId** (Type: string):\n      - **defaultScheme** (Type: boolean):\n  - **startAt** (Type: integer, int64):\n  - **total** (Type: integer, int32):\n  - **maxResults** (Type: integer, int32):\n"

// NewGetPrioritySchemesMCPTool creates the MCP Tool instance for GetPrioritySchemes
func NewGetPrioritySchemesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPrioritySchemes",
		"Get all priority schemes - Returns all priority schemes. All project keys associated with the priority scheme will only be returned if additional query parameter is provided <code>expand=schemes.projectKeys</code>",
		[]byte(GetPrioritySchemesInputSchema),
	)
}

// GetPrioritySchemesHandler is the handler function for the GetPrioritySchemes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPrioritySchemesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/priorityschemes", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPrioritySchemes")
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
