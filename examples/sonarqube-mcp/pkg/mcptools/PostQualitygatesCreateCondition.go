package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostQualitygatesCreateCondition tool
const PostQualitygatesCreateConditionInputSchema = "{\n  \"properties\": {\n    \"error\": {\n      \"description\": \"Condition error threshold\",\n      \"maxLength\": 64,\n      \"type\": \"string\"\n    },\n    \"gateName\": {\n      \"description\": \"Name of the quality gate\",\n      \"type\": \"string\"\n    },\n    \"metric\": {\n      \"description\": \"Condition metric. Only metric of the following types are allowed: - INT - MILLISEC - RATING - WORK_DUR - FLOAT - PERCENT - LEVELFollowing metrics are forbidden: - new_security_hotspots - security_hotspots - alert_status\",\n      \"type\": \"string\"\n    },\n    \"op\": {\n      \"description\": \"Condition operator: - LT = is lower than - GT = is greater than\",\n      \"enum\": [\n        \"LT\",\n        \"GT\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"error\",\n    \"gateName\",\n    \"metric\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the PostQualitygatesCreateCondition tool (Status: 200, Content-Type: application/json)
const PostQualitygatesCreateConditionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewPostQualitygatesCreateConditionMCPTool creates the MCP Tool instance for PostQualitygatesCreateCondition
func NewPostQualitygatesCreateConditionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostQualitygatesCreateCondition",
		"Add a new condition to a quality gate. - Add a new condition to a quality gate. Parameter 'gateName' must be provided. Requires the 'Administer Quality Gates' permission.",
		[]byte(PostQualitygatesCreateConditionInputSchema),
	)
}

// PostQualitygatesCreateConditionHandler is the handler function for the PostQualitygatesCreateCondition tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostQualitygatesCreateConditionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/qualitygates/create_condition", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostQualitygatesCreateCondition")
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
