package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAutoComplete tool
const GetAutoCompleteInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAutoComplete tool (Status: 200, Content-Type: application/json)
const GetAutoCompleteResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The auto complete data required for JQL searches.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **visibleFunctionNames** (Type: array):\n      - Example: '[{\"displayName\":\"currentLogin()\",\"types\":[\"java.util.Date\"],\"value\":\"currentLogin()\"},{\"displayName\":\"currentUser()\",\"types\":[\"com.atlassian.crowd.embedded.api.User\"],\"value\":\"currentUser()\"}]'\n    - **Items** (Type: string):\n        - Example: '[{\"value\":\"currentLogin()\",\"displayName\":\"currentLogin()\",\"types\":[\"java.util.Date\"]},{\"value\":\"currentUser()\",\"displayName\":\"currentUser()\",\"types\":[\"com.atlassian.crowd.embedded.api.User\"]}]'\n  - **jqlReservedWords** (Type: array):\n      - Example: '[\"empty\",\"and\",\"or\",\"in\",\"distinct\"]'\n    - **Items** (Type: string):\n        - Example: '[\"empty\",\"and\",\"or\",\"in\",\"distinct\"]'\n  - **visibleFieldNames** (Type: array):\n      - Example: '[{\"auto\":\"true\",\"displayName\":\"affectedVersion\",\"operators\":[\"=\",\"!=\",\"in\",\"not in\",\"is\",\"is not\",\"\\u003c\",\"\\u003c=\",\"\\u003e\",\"\\u003e=\"],\"orderable\":\"true\",\"searchable\":\"true\",\"types\":[\"com.atlassian.crowd.embedded.api.User\"],\"value\":\"affectedVersion\"},{\"auto\":\"true\",\"displayName\":\"assignee\",\"operators\":[\"!=\",\"was not in\",\"not in\",\"was not\",\"is\",\"was in\",\"was\",\"=\",\"in\",\"changed\",\"is not\"],\"orderable\":\"true\",\"searchable\":\"true\",\"types\":[\"com.atlassian.crowd.embedded.api.User\"],\"value\":\"assignee\"}]'\n    - **Items** (Type: string):\n        - Example: '[{\"value\":\"affectedVersion\",\"displayName\":\"affectedVersion\",\"auto\":\"true\",\"orderable\":\"true\",\"searchable\":\"true\",\"operators\":[\"=\",\"!=\",\"in\",\"not in\",\"is\",\"is not\",\"<\",\"<=\",\">\",\">=\"],\"types\":[\"com.atlassian.crowd.embedded.api.User\"]},{\"value\":\"assignee\",\"displayName\":\"assignee\",\"auto\":\"true\",\"orderable\":\"true\",\"searchable\":\"true\",\"operators\":[\"!=\",\"was not in\",\"not in\",\"was not\",\"is\",\"was in\",\"was\",\"=\",\"in\",\"changed\",\"is not\"],\"types\":[\"com.atlassian.crowd.embedded.api.User\"]}]'\n"

// NewGetAutoCompleteMCPTool creates the MCP Tool instance for GetAutoComplete
func NewGetAutoCompleteMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAutoComplete",
		"Get auto complete data for JQL searches - Returns the auto complete data required for JQL searches",
		[]byte(GetAutoCompleteInputSchema),
	)
}

// GetAutoCompleteHandler is the handler function for the GetAutoComplete tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAutoCompleteHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/jql/autocompletedata", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAutoComplete"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
