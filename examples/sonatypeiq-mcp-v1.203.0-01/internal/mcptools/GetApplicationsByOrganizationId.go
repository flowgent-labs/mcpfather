package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetApplicationsByOrganizationId tool
const GetApplicationsByOrganizationIdInputSchema = "{\n  \"properties\": {\n    \"organizationId\": {\n      \"description\": \"Enter the organizationId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetApplicationsByOrganizationId tool (Status: 200, Content-Type: application/json)
const GetApplicationsByOrganizationIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of all applications found under the organizationId provided.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applications** (Type: array):\n    - **Items** (Type: object):\n      - **publicId** (Type: string):\n      - **applicationTags** (Type: array):\n        - **Items** (Type: object):\n          - **applicationId** (Type: string):\n          - **id** (Type: string):\n          - **tagId** (Type: string):\n      - **contactUserName** (Type: string):\n      - **id** (Type: string):\n      - **name** (Type: string):\n      - **organizationId** (Type: string):\n"

// NewGetApplicationsByOrganizationIdMCPTool creates the MCP Tool instance for GetApplicationsByOrganizationId
func NewGetApplicationsByOrganizationIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplicationsByOrganizationId",
		"Use this method to retrieve application details for all applications under the organizationId provided.\n\nPermissions required: View IQ Elements",
		[]byte(GetApplicationsByOrganizationIdInputSchema),
	)
}

// GetApplicationsByOrganizationIdHandler is the handler function for the GetApplicationsByOrganizationId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicationsByOrganizationIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications/organization/{organizationId}", args, []string{"organizationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplicationsByOrganizationId")
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
