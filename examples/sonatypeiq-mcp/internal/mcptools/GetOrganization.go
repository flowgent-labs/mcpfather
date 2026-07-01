package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetOrganization tool
const GetOrganizationInputSchema = "{\n  \"properties\": {\n    \"organizationId\": {\n      \"description\": \"Enter the organization id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetOrganization tool (Status: 200, Content-Type: application/json)
const GetOrganizationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details for the specified  organization including organization id, organization name, parent organization id, and its associated tags with additional details.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n  - **parentOrganizationId** (Type: string):\n  - **tags** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n      - **name** (Type: string):\n      - **color** (Type: string):\n          - Enum: ['white', 'grey', 'black', 'green', 'yellow', 'orange', 'red', 'blue', 'light-red', 'light-green', 'light-blue', 'light-purple', 'dark-red', 'dark-green', 'dark-blue', 'dark-purple']\n      - **description** (Type: string):\n  - **id** (Type: string):\n"

// NewGetOrganizationMCPTool creates the MCP Tool instance for GetOrganization
func NewGetOrganizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetOrganization",
		"Use this method to retrieve the details of an organization by providing the organization id.\n\nPermissions required: View IQ Elements",
		[]byte(GetOrganizationInputSchema),
	)
}

// GetOrganizationHandler is the handler function for the GetOrganization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetOrganizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/organizations/{organizationId}", args, []string{"organizationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetOrganization")
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
