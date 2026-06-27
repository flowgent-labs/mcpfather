package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetResolution tool
const GetResolutionInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"A String containing the resolution id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetResolution tool (Status: 200, Content-Type: application/json)
const GetResolutionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a Jira issue resolution.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: string):\n      - Example: '1'\n  - **name** (Type: string):\n      - Example: 'Fixed'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/resolution/1'\n  - **description** (Type: string):\n      - Example: 'Fixed'\n  - **iconUrl** (Type: string):\n      - Example: 'http://www.example.com/jira/images/icons/statuses/resolved.png'\n"

// NewGetResolutionMCPTool creates the MCP Tool instance for GetResolution
func NewGetResolutionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetResolution",
		"Get a resolution by ID - Returns a resolution.",
		[]byte(GetResolutionInputSchema),
	)
}

// GetResolutionHandler is the handler function for the GetResolution tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetResolutionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/resolution/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetResolution"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
