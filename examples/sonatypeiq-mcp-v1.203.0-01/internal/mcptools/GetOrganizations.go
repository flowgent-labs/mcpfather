package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetOrganizations tool
const GetOrganizationsInputSchema = "{\n  \"properties\": {\n    \"organizationName\": {\n      \"description\": \"Enter the organization names.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetOrganizations tool (Status: 200, Content-Type: application/json)
const GetOrganizationsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a list of organizations. For each organization the response contains organization id, organization name, parent organization id, and its associated tags with additional details.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **organizations** (Type: array):\n    - **Items** (Type: object):\n      - **tags** (Type: array):\n        - **Items** (Type: object):\n          - **color** (Type: string):\n              - Enum: ['white', 'grey', 'black', 'green', 'yellow', 'orange', 'red', 'blue', 'light-red', 'light-green', 'light-blue', 'light-purple', 'dark-red', 'dark-green', 'dark-blue', 'dark-purple']\n          - **description** (Type: string):\n          - **id** (Type: string):\n          - **name** (Type: string):\n      - **id** (Type: string):\n      - **name** (Type: string):\n      - **parentOrganizationId** (Type: string):\n"

// NewGetOrganizationsMCPTool creates the MCP Tool instance for GetOrganizations
func NewGetOrganizationsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetOrganizations",
		"Use this method to retrieve organizations with names matching those specified or all if not specified.\n\nPermissions required: View IQ Elements",
		[]byte(GetOrganizationsInputSchema),
	)
}

// GetOrganizationsHandler is the handler function for the GetOrganizations tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetOrganizationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/organizations", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetOrganizations")
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
