package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the Update tool
const UpdateInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify the user details to be updated. Any unspecified field will remain unchanged. Username, password, and realm cannot be updated.\",\n      \"properties\": {\n        \"email\": {\n          \"type\": \"string\"\n        },\n        \"firstName\": {\n          \"type\": \"string\"\n        },\n        \"lastName\": {\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"type\": \"string\"\n        },\n        \"realm\": {\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"username\": {\n      \"description\": \"Enter the username.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Update tool (Status: 200, Content-Type: application/json)
const UpdateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> User details updated successfully.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **email** (Type: string):\n  - **firstName** (Type: string):\n  - **lastName** (Type: string):\n  - **password** (Type: string):\n  - **realm** (Type: string):\n  - **username** (Type: string):\n"

// NewUpdateMCPTool creates the MCP Tool instance for Update
func NewUpdateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Update",
		"Use this method to update user details for an existing internal user, by specifying the username.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(UpdateInputSchema),
	)
}

// UpdateHandler is the handler function for the Update tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/users/{username}", args, []string{"username"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Update")
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
