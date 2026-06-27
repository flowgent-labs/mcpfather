package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetOrganizationsByIds tool
const GetOrganizationsByIdsInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"Enter the internal organization IDs.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetOrganizationsByIds tool (Status: 200, Content-Type: application/json)
const GetOrganizationsByIdsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a list of organizations. For each organization the response contains organization id, organization name, and parent organization id.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **organizations** (Type: array):\n    - **Items** (Type: object):\n      - **name** (Type: string):\n      - **parentOrganizationId** (Type: string):\n      - **tags** (Type: array):\n        - **Items** (Type: object):\n          - **color** (Type: string):\n              - Enum: ['white', 'grey', 'black', 'green', 'yellow', 'orange', 'red', 'blue', 'light-red', 'light-green', 'light-blue', 'light-purple', 'dark-red', 'dark-green', 'dark-blue', 'dark-purple']\n          - **description** (Type: string):\n          - **id** (Type: string):\n          - **name** (Type: string):\n      - **id** (Type: string):\n"

// NewGetOrganizationsByIdsMCPTool creates the MCP Tool instance for GetOrganizationsByIds
func NewGetOrganizationsByIdsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetOrganizationsByIds",
		"Use this method to retrieve organizations by their internal IDs.\n\nPermissions required: View IQ Elements",
		[]byte(GetOrganizationsByIdsInputSchema),
	)
}

// GetOrganizationsByIdsHandler is the handler function for the GetOrganizationsByIds tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetOrganizationsByIdsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/organizations/byid", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetOrganizationsByIds")
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
