package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetApplications tool
const GetApplicationsInputSchema = "{\n  \"properties\": {\n    \"includeCategories\": {\n      \"default\": false,\n      \"description\": \"Set this parameter to " + "\x60" + "true" + "\x60" + " to obtain the application tags (application categories) in the response.\",\n      \"type\": \"boolean\"\n    },\n    \"publicId\": {\n      \"description\": \"Enter the applicationId.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetApplications tool (Status: 200, Content-Type: application/json)
const GetApplicationsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns either a list of applications or a list of applications with category tags depending on the " + "\x60" + "includeCategories" + "\x60" + " parameter.\n\n## Response Structure\n\n- Structure (Type: Combinator):\n  - **One Of the following structures**:\n    - **Option 1** (Type: object):\n      - **applications** (Type: array):\n        - **Items** (Type: object):\n          - **applicationTags** (Type: array):\n            - **Items** (Type: object):\n              - **applicationId** (Type: string):\n              - **id** (Type: string):\n              - **tagId** (Type: string):\n          - **contactUserName** (Type: string):\n          - **id** (Type: string):\n          - **name** (Type: string):\n          - **organizationId** (Type: string):\n          - **publicId** (Type: string):\n    - **Option 2** (Type: object):\n      - **applications** (Type: array):\n        - **Items** (Type: object):\n          - **categories** (Type: array):\n            - **Items** (Type: object):\n              - **id** (Type: string):\n              - **name** (Type: string):\n              - **organizationId** (Type: string):\n              - **color** (Type: string):\n              - **description** (Type: string):\n          - **contactUserName** (Type: string):\n          - **id** (Type: string):\n          - **name** (Type: string):\n          - **organizationId** (Type: string):\n          - **publicId** (Type: string):\n"

// NewGetApplicationsMCPTool creates the MCP Tool instance for GetApplications
func NewGetApplicationsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplications",
		"Use this method to retrieve the application details for the applicationId(s) provided.\n\nPermissions required: View IQ Elements",
		[]byte(GetApplicationsInputSchema),
	)
}

// GetApplicationsHandler is the handler function for the GetApplications tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplications")
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
