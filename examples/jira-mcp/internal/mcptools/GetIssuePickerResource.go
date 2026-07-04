package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetIssuePickerResource tool
const GetIssuePickerResourceInputSchema = "{\n  \"properties\": {\n    \"currentIssueKey\": {\n      \"description\": \"the key of the issue in context of which the request is executed\",\n      \"type\": \"string\"\n    },\n    \"currentJQL\": {\n      \"description\": \"the JQL in context of which the request is executed\",\n      \"type\": \"string\"\n    },\n    \"currentProjectId\": {\n      \"description\": \"the id of the project in context of which the request is executed\",\n      \"type\": \"string\"\n    },\n    \"query\": {\n      \"description\": \"the query\",\n      \"type\": \"string\"\n    },\n    \"showSubTaskParent\": {\n      \"description\": \"if set to false and request is executed in context of a subtask, the parent issue will not be included in the auto-completion result, even if it matches the query\",\n      \"type\": \"string\"\n    },\n    \"showSubTasks\": {\n      \"description\": \"if set to false, subtasks will not be included in the list\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetIssuePickerResource tool (Status: 200, Content-Type: application/json)
const GetIssuePickerResourceResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a response containing issue picker resource.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **sections** (Type: array):\n    - **Items** (Type: object):\n      - **issues** (Type: array):\n        - **Items** (Type: object):\n          - **keyHtml** (Type: string):\n              - Example: 'issueKeyHtml'\n          - **summary** (Type: string):\n              - Example: 'summary'\n          - **summaryText** (Type: string):\n              - Example: 'summaryText'\n          - **img** (Type: string):\n              - Example: 'img'\n          - **key** (Type: string):\n              - Example: 'issueKey'\n      - **label** (Type: string):\n          - Example: 'section'\n      - **msg** (Type: string):\n          - Example: 'msg'\n      - **sub** (Type: string):\n          - Example: 'sub'\n      - **id** (Type: string):\n          - Example: 'id'\n"

// NewGetIssuePickerResourceMCPTool creates the MCP Tool instance for GetIssuePickerResource
func NewGetIssuePickerResourceMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssuePickerResource",
		"Get suggested issues for auto-completion - Get issue picker resource",
		[]byte(GetIssuePickerResourceInputSchema),
	)
}

// GetIssuePickerResourceHandler is the handler function for the GetIssuePickerResource tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssuePickerResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issue/picker", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetIssuePickerResource")
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
