package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the CreateBoard tool
const CreateBoardInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Bean which contains board name, type and filter Id.\",\n      \"properties\": {\n        \"filterId\": {\n          \"example\": 10040,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"name\": {\n          \"example\": \"scrum board\",\n          \"type\": \"string\"\n        },\n        \"type\": {\n          \"example\": \"scrum\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreateBoard tool (Status: 201, Content-Type: application/json)
const CreateBoardResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returns the created board.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **name** (Type: string):\n      - Example: 'Scrum Board'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/board/10001'\n  - **type** (Type: string):\n      - Example: 'scrum'\n"

// NewCreateBoardMCPTool creates the MCP Tool instance for CreateBoard
func NewCreateBoardMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateBoard",
		"Create a new board - Creates a new board. Board name, type and filter Id is required.\n- name - Must be less than 255 characters.\n- type - Valid values: scrum, kanban\n- filterId - Id of a filter that the user has permissions to view. Note, if the user does not have the 'Create shared objects' permission and tries to create a shared board, a private board will be created instead (remember that board sharing depends on the filter sharing).\nNote:\n- If you want to create a new project with an associated board, use the JIRA platform REST API. For more information, see the Create project method. The projectTypeKey for software boards must be 'software' and the projectTemplateKey must be either com.pyxis.greenhopper.jira:gh-kanban-template or com.pyxis.greenhopper.jira:gh-scrum-template.\n- You can create a filter using the JIRA REST API. For more information, see the Create filter method.\n- If you do not ORDER BY the Rank field for the filter of your board, you will not be able to reorder issues on the board.",
		[]byte(CreateBoardInputSchema),
	)
}

// CreateBoardHandler is the handler function for the CreateBoard tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateBoardHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/agile/1.0/board", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateBoard"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
