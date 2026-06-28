package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetNotificationScheme tool
const GetNotificationSchemeInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Optional information to be expanded in the response: group, user, projectRole or field.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"The id of the notification scheme to retrieve\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetNotificationScheme tool (Status: 200, Content-Type: application/json)
const GetNotificationSchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full representation of the notification scheme with given id\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/notificationscheme/10100'\n  - **description** (Type: string):\n      - Example: 'description'\n  - **expand** (Type: string):\n  - **id** (Type: integer, int64):\n      - Example: '10100'\n  - **name** (Type: string):\n      - Example: 'notification scheme name'\n  - **notificationSchemeEvents** (Type: object):\n"

// NewGetNotificationSchemeMCPTool creates the MCP Tool instance for GetNotificationScheme
func NewGetNotificationSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetNotificationScheme",
		"Get full notification scheme details - Returns a full representation of the notification scheme for the given id. This resource will return a\nnotification scheme containing a list of events and recipient configured to receive notifications for these events. Consumer\nshould allow events without recipients to appear in response. User accessing\nthe data is required to have permissions to administer at least one project associated with the requested notification scheme.\nNotification recipients can be:\n- current assignee - the value of the notificationType is CurrentAssignee\n- issue reporter - the value of the notificationType is Reporter\n- current user - the value of the notificationType is CurrentUser\n- project lead - the value of the notificationType is ProjectLead\n- component lead - the value of the notificationType is ComponentLead\n- all watchers - the value of the notification type is AllWatchers\n<li>configured user - the value of the notification type is User. Parameter will contain key of the user. Information about the user will be provided\nif <b>user</b> expand parameter is used.\n- configured group - the value of the notification type is Group. Parameter will contain name of the group. Information about the group will be provided\nif <b>group</b> expand parameter is used.\n- configured email address - the value of the notification type is EmailAddress, additionally information about the email will be provided.\n- users or users in groups in the configured custom fields - the value of the notification type is UserCustomField or GroupCustomField. Parameter\nwill contain id of the custom field. Information about the field will be provided if <b>field</b> expand parameter is used.\n- configured project role - the value of the notification type is ProjectRole. Parameter will contain project role id. Information about the project role\nwill be provided if <b>projectRole</b> expand parameter is used.\nPlease see the example for reference.\nThe events can be Jira system events or events configured by administrator. In case of the system events, data about theirs\nids, names and descriptions is provided. In case of custom events, the template event is included as well.",
		[]byte(GetNotificationSchemeInputSchema),
	)
}

// GetNotificationSchemeHandler is the handler function for the GetNotificationScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetNotificationSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/notificationscheme/{id}", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetNotificationScheme")
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
