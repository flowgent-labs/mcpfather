package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetApplication tool
const GetApplicationInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the applicationId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetApplication tool (Status: 200, Content-Type: application/json)
const GetApplicationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of the application corresponding to the applicationId.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: string):\n  - **name** (Type: string):\n  - **organizationId** (Type: string):\n  - **publicId** (Type: string):\n  - **applicationTags** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n      - **tagId** (Type: string):\n      - **applicationId** (Type: string):\n  - **contactUserName** (Type: string):\n"

// NewGetApplicationMCPTool creates the MCP Tool instance for GetApplication
func NewGetApplicationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplication",
		"Use this method to retrieve the application details, by providing the applicationId.\n\nPermissions required: View IQ Elements",
		[]byte(GetApplicationInputSchema),
	)
}

// GetApplicationHandler is the handler function for the GetApplication tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications/{applicationId}", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplication")
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
