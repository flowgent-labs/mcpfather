package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the EvaluateSourceControl tool
const EvaluateSourceControlInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal applicationId. Use the Applications REST API to retrieve the internal applicationId.\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"The request JSON should include the 1. branch name (name of the target branch in the source control repository, 2. stageId (recommended values are 'develop' for feature branches, and 'source' for default branches. Other stageIds that can be used are 'build', 'stage-release', 'release', 'operate' but are not recommended), 3. scanTargets (optional, specify one or more paths inside the repository. If not specified, the entire repository will be evaluated by default). Ensure that the repository paths are not relative and do not contain '../' or '..\\\\'.\",\n      \"properties\": {\n        \"branchName\": {\n          \"type\": \"string\"\n        },\n        \"scanTargets\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"stageId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the EvaluateSourceControl tool (Status: 200, Content-Type: application/json)
const EvaluateSourceControlResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains statusUrl. Use this statusUrl to check the evaluation status using the GET method (step 2 of the evaluation process). \n\n## Response Structure\n\n- Structure (Type: object):\n  - **statusUrl** (Type: string):\n"

// NewEvaluateSourceControlMCPTool creates the MCP Tool instance for EvaluateSourceControl
func NewEvaluateSourceControlMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"EvaluateSourceControl",
		"Use this method to request a source control evaluation for a specific application. This is step 1 of the 2 step source control evaluation process. \n\nPermissions Required: Evaluate Applications",
		[]byte(EvaluateSourceControlInputSchema),
	)
}

// EvaluateSourceControlHandler is the handler function for the EvaluateSourceControl tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func EvaluateSourceControlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/evaluation/applications/{applicationId}/sourceControlEvaluation", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "EvaluateSourceControl")
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
