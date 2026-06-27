package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the CloneApplication tool
const CloneApplicationInputSchema = "{\n  \"properties\": {\n    \"clonedApplicationName\": {\n      \"description\": \"Enter the application name for the new cloned application.\",\n      \"type\": \"string\"\n    },\n    \"clonedApplicationPublicId\": {\n      \"description\": \"Enter the applicationPublicId for the cloned application.\",\n      \"type\": \"string\"\n    },\n    \"sourceApplicationId\": {\n      \"description\": \"Enter the applicationId for the application to be cloned.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"sourceApplicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CloneApplication tool (Status: 200, Content-Type: application/json)
const CloneApplicationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains application details of the cloned application.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **organizationId** (Type: string):\n  - **publicId** (Type: string):\n  - **applicationTags** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n      - **tagId** (Type: string):\n      - **applicationId** (Type: string):\n  - **contactUserName** (Type: string):\n  - **id** (Type: string):\n  - **name** (Type: string):\n"

// NewCloneApplicationMCPTool creates the MCP Tool instance for CloneApplication
func NewCloneApplicationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CloneApplication",
		"Use this method to clone an existing application.\n\nPermissions required: Add Application (on the parent organization)",
		[]byte(CloneApplicationInputSchema),
	)
}

// CloneApplicationHandler is the handler function for the CloneApplication tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CloneApplicationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/applications/{sourceApplicationId}/clone", args, []string{"sourceApplicationId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CloneApplication"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
