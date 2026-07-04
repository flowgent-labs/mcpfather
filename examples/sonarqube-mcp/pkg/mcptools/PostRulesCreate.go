package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostRulesCreate tool
const PostRulesCreateInputSchema = "{\n  \"properties\": {\n    \"cleanCodeAttribute\": {\n      \"description\": \"Clean code attribute\",\n      \"enum\": [\n        \"CONVENTIONAL\",\n        \"FORMATTED\",\n        \"IDENTIFIABLE\",\n        \"CLEAR\",\n        \"COMPLETE\",\n        \"EFFICIENT\",\n        \"LOGICAL\",\n        \"DISTINCT\",\n        \"FOCUSED\",\n        \"MODULAR\",\n        \"TESTED\",\n        \"LAWFUL\",\n        \"RESPECTFUL\",\n        \"TRUSTWORTHY\"\n      ],\n      \"type\": \"string\"\n    },\n    \"customKey\": {\n      \"description\": \"Key of the custom rule\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    },\n    \"impacts\": {\n      \"description\": \"Impacts as semi-colon list of \\u003csoftware_quality\\u003e=\\u003cseverity\\u003e\",\n      \"type\": \"string\"\n    },\n    \"markdownDescription\": {\n      \"description\": \"Rule description in markdown format\",\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"Rule name\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    },\n    \"params\": {\n      \"description\": \"Parameters as semi-colon list of \\u003ckey\\u003e=\\u003cvalue\\u003e\",\n      \"type\": \"string\"\n    },\n    \"preventReactivation\": {\n      \"description\": \"If set to true and if the rule has been deactivated (status 'REMOVED'), a status 409 will be returned\",\n      \"enum\": [\n        \"true\",\n        \"false\",\n        \"yes\",\n        \"no\"\n      ],\n      \"type\": \"string\"\n    },\n    \"severity\": {\n      \"description\": \"Rule severity\",\n      \"enum\": [\n        \"INFO\",\n        \"MINOR\",\n        \"MAJOR\",\n        \"CRITICAL\",\n        \"BLOCKER\"\n      ],\n      \"type\": \"string\"\n    },\n    \"status\": {\n      \"description\": \"Rule status\",\n      \"enum\": [\n        \"BETA\",\n        \"DEPRECATED\",\n        \"READY\"\n      ],\n      \"type\": \"string\"\n    },\n    \"templateKey\": {\n      \"description\": \"Key of the template rule in order to create a custom rule\",\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"description\": \"Rule type\",\n      \"enum\": [\n        \"CODE_SMELL\",\n        \"BUG\",\n        \"VULNERABILITY\",\n        \"SECURITY_HOTSPOT\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"customKey\",\n    \"markdownDescription\",\n    \"name\",\n    \"templateKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PostRulesCreate tool (Status: 200, Content-Type: application/json)
const PostRulesCreateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewPostRulesCreateMCPTool creates the MCP Tool instance for PostRulesCreate
func NewPostRulesCreateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostRulesCreate",
		"Create a custom rule. - Create a custom rule. Requires the 'Administer Quality Profiles' permission",
		[]byte(PostRulesCreateInputSchema),
	)
}

// PostRulesCreateHandler is the handler function for the PostRulesCreate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostRulesCreateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/rules/create", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostRulesCreate")
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
