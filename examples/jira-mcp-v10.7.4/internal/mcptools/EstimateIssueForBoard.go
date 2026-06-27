package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the EstimateIssueForBoard tool
const EstimateIssueForBoardInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"body\": {\n      \"description\": \"Bean that contains value of a new estimation.\",\n      \"properties\": {\n        \"value\": {\n          \"example\": \"8.0\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"issueIdOrKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the EstimateIssueForBoard tool (Status: 200, Content-Type: application/json)
const EstimateIssueForBoardResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the estimation of the issue and a fieldId of the field that is used for it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **fieldId** (Type: string):\n      - Example: 'customfield_12532'\n  - **value** (Type: object):\n      - Example: '8'\n"

// NewEstimateIssueForBoardMCPTool creates the MCP Tool instance for EstimateIssueForBoard
func NewEstimateIssueForBoardMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"EstimateIssueForBoard",
		"Update the estimation of an issue for a board - Updates the estimation of the issue. boardId param is required. This param determines which field will be updated on a issue.\nNote that this resource changes the estimation field of the issue regardless of appearance the field on the screen.\nOriginal time tracking estimation field accepts estimation in formats like \"1w\", \"2d\", \"3h\", \"20m\" or number which represent number of minutes.\nHowever, internally the field stores and returns the estimation as a number of seconds.\nThe field used for estimation on the given board can be obtained from <a href=\"#agile/1.0/board-getConfiguration\">board configuration resource</a>.\nMore information about the field are returned by edit meta resource or field resource.",
		[]byte(EstimateIssueForBoardInputSchema),
	)
}

// EstimateIssueForBoardHandler is the handler function for the EstimateIssueForBoard tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func EstimateIssueForBoardHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/agile/1.0/issue/{issueIdOrKey}/estimation", args, []string{"issueIdOrKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "EstimateIssueForBoard"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
