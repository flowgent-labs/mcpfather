package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueEstimationForBoard tool
const GetIssueEstimationForBoardInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"issueIdOrKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueEstimationForBoard tool (Status: 200, Content-Type: application/json)
const GetIssueEstimationForBoardResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the estimation of the issue and a fieldId of the field that is used for it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **value** (Type: object):\n      - Example: '8'\n  - **fieldId** (Type: string):\n      - Example: 'customfield_12532'\n"

// NewGetIssueEstimationForBoardMCPTool creates the MCP Tool instance for GetIssueEstimationForBoard
func NewGetIssueEstimationForBoardMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueEstimationForBoard",
		"Get the estimation of an issue for a board - Returns the estimation of the issue and a fieldId of the field that is used for it.\nOriginal time internally stores and returns the estimation as a number of seconds.\nThe field used for estimation on the given board can be obtained from board configuration resource.\nMore information about the field are returned by edit meta resource or field resource.",
		[]byte(GetIssueEstimationForBoardInputSchema),
	)
}

// GetIssueEstimationForBoardHandler is the handler function for the GetIssueEstimationForBoard tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueEstimationForBoardHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/issue/{issueIdOrKey}/estimation", args, []string{"issueIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueEstimationForBoard"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
