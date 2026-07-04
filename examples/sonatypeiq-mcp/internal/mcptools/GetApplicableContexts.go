package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetApplicableContexts tool
const GetApplicableContextsInputSchema = "{\n  \"properties\": {\n    \"labelId\": {\n      \"description\": \"Enter the labelId\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the ownerId\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the ownerType.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"labelId\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetApplicableContexts tool (Status: 200, Content-Type: application/json)
const GetApplicableContextsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "id" + "\x60" + " is the id of the selected owner.</li><li>" + "\x60" + "name" + "\x60" + " is the name of the selected owner.</li><li>" + "\x60" + "type" + "\x60" + " is the type of the selected owner e.g. application, organization or repository.</li><li>" + "\x60" + "children" + "\x60" + " is an array of the child owners in the hierarchy.</li>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n  - **type** (Type: string):\n      - Enum: ['application', 'organization', 'repository_container', 'repository_manager', 'repository', 'global']\n  - **children** (Type: array):\n    - **[cyclic reference]**\n  - **id** (Type: string):\n"

// NewGetApplicableContextsMCPTool creates the MCP Tool instance for GetApplicableContexts
func NewGetApplicableContextsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplicableContexts",
		"Use this method to retrieve the hierarchy of owners (applications, organizations, repositories) in which the label can be applied.\n\nPermissions required: Edit IQ Elements",
		[]byte(GetApplicableContextsInputSchema),
	)
}

// GetApplicableContextsHandler is the handler function for the GetApplicableContexts tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicableContextsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/labels/{ownerType}/{ownerId}/applicable/context/{labelId}", args, []string{"labelId", "ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplicableContexts")
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
