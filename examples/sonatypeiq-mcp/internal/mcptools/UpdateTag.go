package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the UpdateTag tool
const UpdateTagInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify the id (application category id) and id of the organization that owns this  application category, to update the name, description and color.\",\n      \"properties\": {\n        \"color\": {\n          \"type\": \"string\"\n        },\n        \"description\": {\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"organizationId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"organizationId\": {\n      \"description\": \"The organizationId assigned by IQ Server.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateTag tool (Status: 200, Content-Type: application/json)
const UpdateTagResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful update echoing the updated application category details.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **organizationId** (Type: string):\n  - **color** (Type: string):\n  - **description** (Type: string):\n  - **id** (Type: string):\n  - **name** (Type: string):\n"

// NewUpdateTagMCPTool creates the MCP Tool instance for UpdateTag
func NewUpdateTagMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateTag",
		"Grouping applications with similar characteristics into categories makes policy management easier. You can then create a policy that applies to a specific category. Use this method to update an existing application category.",
		[]byte(UpdateTagInputSchema),
	)
}

// UpdateTagHandler is the handler function for the UpdateTag tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateTagHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/applicationCategories/organization/{organizationId}", args, []string{"organizationId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateTag")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
