package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetNotificationScheme1 tool
const GetNotificationScheme1InputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Optional information to be expanded in the response: group, user, projectRole or field.\",\n      \"type\": \"string\"\n    },\n    \"projectKeyOrId\": {\n      \"description\": \"Key or id of the project\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectKeyOrId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetNotificationScheme1 tool (Status: 200, Content-Type: application/json)
const GetNotificationScheme1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full representation of the notification scheme with given id\n\n## Response Structure\n\n- Structure (Type: object):\n  - **expand** (Type: string):\n  - **id** (Type: integer, int64):\n      - Example: '10100'\n  - **name** (Type: string):\n      - Example: 'notification scheme name'\n  - **notificationSchemeEvents** (Type: object):\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/notificationscheme/10100'\n  - **description** (Type: string):\n      - Example: 'description'\n"

// NewGetNotificationScheme1MCPTool creates the MCP Tool instance for GetNotificationScheme1
func NewGetNotificationScheme1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetNotificationScheme1",
		"Get notification scheme associated with the project - Gets a notification scheme associated with the project. Follow the documentation of /notificationscheme/{id} resource for all details about returned value.",
		[]byte(GetNotificationScheme1InputSchema),
	)
}

// GetNotificationScheme1Handler is the handler function for the GetNotificationScheme1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetNotificationScheme1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/project/{projectKeyOrId}/notificationscheme", args, []string{"projectKeyOrId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetNotificationScheme1")
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
