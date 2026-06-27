package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddOrganization tool
const AddOrganizationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON should include the name of the organization (should be unique), name of the parent organization and tags containing additional organization details. If the parent organization is not specified, this organization will be created under the root organization. Tags represent identifying characteristics of an application. They are created at the organization level and then applied to applications under the organization. The tags can be used to decide which applications will be evaluated against a selected policy.\",\n      \"properties\": {\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"parentOrganizationId\": {\n          \"type\": \"string\"\n        },\n        \"tags\": {\n          \"items\": {\n            \"properties\": {\n              \"color\": {\n                \"enum\": [\n                  \"white\",\n                  \"grey\",\n                  \"black\",\n                  \"green\",\n                  \"yellow\",\n                  \"orange\",\n                  \"red\",\n                  \"blue\",\n                  \"light-red\",\n                  \"light-green\",\n                  \"light-blue\",\n                  \"light-purple\",\n                  \"dark-red\",\n                  \"dark-green\",\n                  \"dark-blue\",\n                  \"dark-purple\"\n                ],\n                \"type\": \"string\"\n              },\n              \"description\": {\n                \"type\": \"string\"\n              },\n              \"id\": {\n                \"type\": \"string\"\n              },\n              \"name\": {\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the AddOrganization tool (Status: 200, Content-Type: application/json)
const AddOrganizationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the assigned organization id and all other organization details specified.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n  - **parentOrganizationId** (Type: string):\n  - **tags** (Type: array):\n    - **Items** (Type: object):\n      - **description** (Type: string):\n      - **id** (Type: string):\n      - **name** (Type: string):\n      - **color** (Type: string):\n          - Enum: ['white', 'grey', 'black', 'green', 'yellow', 'orange', 'red', 'blue', 'light-red', 'light-green', 'light-blue', 'light-purple', 'dark-red', 'dark-green', 'dark-blue', 'dark-purple']\n  - **id** (Type: string):\n"

// NewAddOrganizationMCPTool creates the MCP Tool instance for AddOrganization
func NewAddOrganizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddOrganization",
		"Use this method to add a new organization.\n\nPermissions required: Edit IQ Elements",
		[]byte(AddOrganizationInputSchema),
	)
}

// AddOrganizationHandler is the handler function for the AddOrganization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddOrganizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/organizations", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddOrganization")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
