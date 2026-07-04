package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetRulesList tool
const GetRulesListInputSchema = "{\n  \"properties\": {\n    \"asc\": {\n      \"description\": \"Ascending sort\",\n      \"enum\": [\n        \"true\",\n        \"false\",\n        \"yes\",\n        \"no\"\n      ],\n      \"type\": \"string\"\n    },\n    \"available_since\": {\n      \"description\": \"Filter rules available since the given date. If no value is provided, all rules are returned. Format is yyyy-MM-dd.\",\n      \"type\": \"string\"\n    },\n    \"p\": {\n      \"default\": \"1\",\n      \"description\": \"1-based page number\",\n      \"type\": \"string\"\n    },\n    \"ps\": {\n      \"description\": \"Page size. Must be greater than 0 and less than 500\",\n      \"maximum\": 500,\n      \"type\": \"integer\"\n    },\n    \"qprofile\": {\n      \"description\": \"Filter rules that are activated in the given quality profile.\",\n      \"type\": \"string\"\n    },\n    \"s\": {\n      \"description\": \"Sort field\",\n      \"enum\": [\n        \"createdAt\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetRulesList tool (Status: 200, Content-Type: application/json)
const GetRulesListResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetRulesListMCPTool creates the MCP Tool instance for GetRulesList
func NewGetRulesListMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRulesList",
		"List rules, excluding the external rules and the rules with status REMOVED. - List rules, excluding the external rules and the rules with status REMOVED. If a quality profile is provided, it returns only the rules activated in this quality profile. If no quality profile is provided, it returns all the rules available to the user.",
		[]byte(GetRulesListInputSchema),
	)
}

// GetRulesListHandler is the handler function for the GetRulesList tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRulesListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/rules/list", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRulesList")
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
