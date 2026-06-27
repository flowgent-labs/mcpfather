package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetAppliedTags tool
const GetAppliedTagsInputSchema = "{\n  \"properties\": {\n    \"organizationId\": {\n      \"description\": \"The organizationId assigned by IQ Server.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAppliedTags tool (Status: 200, Content-Type: application/json)
const GetAppliedTagsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Get all application categories or tags that can be applied to an application, belonging to the organization specified by the organization id.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationTagsByOwner** (Type: array):\n    - **Items** (Type: object):\n      - **applicationTags** (Type: array):\n        - **Items** (Type: object):\n          - **tagId** (Type: string):\n          - **applicationId** (Type: string):\n          - **id** (Type: string):\n      - **ownerId** (Type: string):\n      - **ownerName** (Type: string):\n      - **ownerType** (Type: string):\n          - Enum: ['application', 'organization', 'repository_container', 'repository_manager', 'repository', 'global']\n"

// NewGetAppliedTagsMCPTool creates the MCP Tool instance for GetAppliedTags
func NewGetAppliedTagsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAppliedTags",
		"Grouping applications with similar characteristics into categories makes policy management easier. You can then create a policy that applies to a specific category. Use this method to retrieve a list of application categories that can be applied to applications in this organization.",
		[]byte(GetAppliedTagsInputSchema),
	)
}

// GetAppliedTagsHandler is the handler function for the GetAppliedTags tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAppliedTagsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applicationCategories/organization/{organizationId}/applied", args, []string{"organizationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAppliedTags")
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
