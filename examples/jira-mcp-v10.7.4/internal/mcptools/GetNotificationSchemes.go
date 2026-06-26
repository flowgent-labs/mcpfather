package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetNotificationSchemes tool
const GetNotificationSchemesInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Optional information to be expanded in the response: group, user, projectRole or field.\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of notification schemes to return (max 50).\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"startAt\": {\n      \"description\": \"The index of the first notification scheme to return (0 based).\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetNotificationSchemes tool (Status: 200, Content-Type: application/json)
const GetNotificationSchemesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Paginated list of notification schemes to which the user has permissions.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **isLast** (Type: boolean):\n  - **maxResults** (Type: integer, int32):\n  - **nextPage** (Type: string, uri):\n  - **self** (Type: string, uri):\n  - **startAt** (Type: integer, int64):\n  - **total** (Type: integer, int64):\n  - **values** (Type: array):\n    - **Items** (Type: object):\n"

// NewGetNotificationSchemesMCPTool creates the MCP Tool instance for GetNotificationSchemes
func NewGetNotificationSchemesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetNotificationSchemes",
		"Get paginated notification schemes - Returns a paginated list of notification schemes. In order to access notification scheme, the calling user is\nrequired to have permissions to administer at least one project associated with the requested notification scheme. Each scheme contains\na list of events and recipient configured to receive notifications for these events. Consumer should allow events without recipients to appear in response.\nThe list is ordered by the scheme's name.\nFollow the documentation of /notificationscheme/{id} resource for all details about returned value.\n",
		[]byte(GetNotificationSchemesInputSchema),
	)
}

// GetNotificationSchemesHandler is the handler function for the GetNotificationSchemes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetNotificationSchemesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/notificationscheme", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetNotificationSchemes"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
