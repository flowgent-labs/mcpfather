package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetFieldAutoCompleteForQueryString tool
const GetFieldAutoCompleteForQueryStringInputSchema = "{\n  \"properties\": {\n    \"fieldName\": {\n      \"description\": \"The field name for which the suggestions are generated.\",\n      \"type\": \"string\"\n    },\n    \"fieldValue\": {\n      \"description\": \"The portion of the field value that has already been provided by the user.\",\n      \"type\": \"string\"\n    },\n    \"predicateName\": {\n      \"description\": \"The predicate for which the suggestions are generated. Suggestions are generated only for: \\\"by\\\", \\\"from\\\" and \\\"to\\\".\",\n      \"type\": \"string\"\n    },\n    \"predicateValue\": {\n      \"description\": \"The portion of the predicate value that has already been provided by the user.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetFieldAutoCompleteForQueryString tool (Status: 200, Content-Type: application/json)
const GetFieldAutoCompleteForQueryStringResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The autocompletion suggestions for JQL search.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewGetFieldAutoCompleteForQueryStringMCPTool creates the MCP Tool instance for GetFieldAutoCompleteForQueryString
func NewGetFieldAutoCompleteForQueryStringMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetFieldAutoCompleteForQueryString",
		"Get auto complete suggestions for JQL search - Returns auto complete suggestions for JQL search",
		[]byte(GetFieldAutoCompleteForQueryStringInputSchema),
	)
}

// GetFieldAutoCompleteForQueryStringHandler is the handler function for the GetFieldAutoCompleteForQueryString tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetFieldAutoCompleteForQueryStringHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/jql/autocompletedata/suggestions", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetFieldAutoCompleteForQueryString"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
