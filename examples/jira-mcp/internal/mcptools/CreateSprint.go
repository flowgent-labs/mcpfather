package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateSprint tool
const CreateSprintInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The sprint to create.\",\n      \"properties\": {\n        \"autoStartStop\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"endDate\": {\n          \"example\": \"2015-04-20T01:22:00.000+10:00\",\n          \"type\": \"string\"\n        },\n        \"goal\": {\n          \"example\": \"Goal for the sprint\",\n          \"type\": \"string\"\n        },\n        \"incompleteIssuesDestinationId\": {\n          \"example\": 10001,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"name\": {\n          \"example\": \"Sprint 1\",\n          \"type\": \"string\"\n        },\n        \"originBoardId\": {\n          \"example\": 5,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"startDate\": {\n          \"example\": \"2015-04-11T15:22:00.000+10:00\",\n          \"type\": \"string\"\n        },\n        \"synced\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"userProfileTimeZone\": {\n          \"example\": \"Australia/Sydney\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateSprint tool (Status: 201, Content-Type: application/json)
const CreateSprintResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returns the created sprint.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **incompleteIssuesDestinationId** (Type: integer, int64):\n      - Example: '10001'\n  - **name** (Type: string):\n      - Example: 'Sprint 1'\n  - **originBoardId** (Type: integer, int64):\n      - Example: '5'\n  - **startDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **activatedDate** (Type: string):\n      - Example: '2015-04-11T15:22:00.000+10:00'\n  - **completeDate** (Type: string):\n      - Example: '2015-04-20T11:04:00.000+10:00'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/sprint/10001'\n  - **autoStartStop** (Type: boolean):\n      - Example: 'true'\n  - **goal** (Type: string):\n      - Example: 'Goal for the sprint'\n  - **synced** (Type: boolean):\n      - Example: 'true'\n  - **endDate** (Type: string):\n      - Example: '2015-04-20T01:22:00.000+10:00'\n  - **state** (Type: string):\n      - Example: 'active'\n"

// NewCreateSprintMCPTool creates the MCP Tool instance for CreateSprint
func NewCreateSprintMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateSprint",
		"Create a future sprint - Creates a future sprint. Sprint name and origin board id are required. Start and end date are optional. Notes: The sprint name is trimmed. Only Jira administrators can create synced sprints.",
		[]byte(CreateSprintInputSchema),
	)
}

// CreateSprintHandler is the handler function for the CreateSprint tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateSprintHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/sprint", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateSprint")
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
