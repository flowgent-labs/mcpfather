package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the EvaluateComponents tool
const EvaluateComponentsInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal applicationId. Use the Applications REST API to retrieve the internal applicationId.\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"The request JSON should contain component coordinates or the hash (SHA1) for each component. You can provide the packageURL instead of component information or hash.\",\n      \"properties\": {\n        \"components\": {\n          \"items\": {\n            \"properties\": {\n              \"componentIdentifier\": {\n                \"properties\": {\n                  \"coordinates\": {\n                    \"additionalProperties\": {\n                      \"type\": \"string\"\n                    },\n                    \"type\": \"object\"\n                  },\n                  \"format\": {\n                    \"type\": \"string\"\n                  }\n                },\n                \"type\": \"object\"\n              },\n              \"displayName\": {\n                \"type\": \"string\"\n              },\n              \"hash\": {\n                \"type\": \"string\"\n              },\n              \"originalPurl\": {\n                \"type\": \"string\"\n              },\n              \"packageUrl\": {\n                \"type\": \"string\"\n              },\n              \"proprietary\": {\n                \"type\": \"boolean\"\n              },\n              \"sha256\": {\n                \"type\": \"string\"\n              },\n              \"thirdParty\": {\n                \"type\": \"boolean\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the EvaluateComponents tool (Status: 200, Content-Type: application/json)
const EvaluateComponentsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The JSON response contains resultId that will be assigned to the evaluation results, timestamp when the component evaluation was requested, the applicationId of the component and the results URL. The resultId obtained from here can be used to retrieve the evaluation result using the REST API or the result URL can be used in cURL. \n\n## Response Structure\n\n- Structure (Type: object):\n  - **submittedDate** (Type: string, date-time):\n  - **applicationId** (Type: string):\n  - **resultId** (Type: string):\n  - **resultsUrl** (Type: string):\n"

// NewEvaluateComponentsMCPTool creates the MCP Tool instance for EvaluateComponents
func NewEvaluateComponentsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"EvaluateComponents",
		"Use this method to request a component evaluation. This is step 1 of the 2 step policy evaluation for components process.\n\nPermissions Required: Evaluate Components",
		[]byte(EvaluateComponentsInputSchema),
	)
}

// EvaluateComponentsHandler is the handler function for the EvaluateComponents tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func EvaluateComponentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/evaluation/applications/{applicationId}", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "EvaluateComponents")
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
