package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the FindUsersForPicker tool
const FindUsersForPickerInputSchema = "{\n  \"properties\": {\n    \"exclude\": {\n      \"description\": \"List of users to be excluded from the search results\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of users to return (defaults to 50). The maximum allowed value is 100 (The combination of maxResults and startAt is limited to the first 100 results). If you specify a value that is higher than this number, your search results will be truncated. If you send a request with startAt=98 and maxResults=20, it will only return 2 users.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"query\": {\n      \"description\": \"A string used to search username, Name or e-mail address\",\n      \"type\": \"string\"\n    },\n    \"showAvatar\": {\n      \"description\": \"If true, then avatars are included in the results\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the FindUsersForPicker tool (Status: 200, Content-Type: application/json)
const FindUsersForPickerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of users matching query with highlighting.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewFindUsersForPickerMCPTool creates the MCP Tool instance for FindUsersForPicker
func NewFindUsersForPickerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FindUsersForPicker",
		"Find users for picker by query - Returns a list of users matching query with highlighting.",
		[]byte(FindUsersForPickerInputSchema),
	)
}

// FindUsersForPickerHandler is the handler function for the FindUsersForPicker tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FindUsersForPickerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/picker", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "FindUsersForPicker"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
