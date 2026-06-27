package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the TestCrowdConfiguration tool
const TestCrowdConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"To test an existing configuration, the request body is not required.\\n\\nTo test a new configuration, provide the " + "\x60" + "serverURl" + "\x60" + ", " + "\x60" + "applicationName" + "\x60" + ", and " + "\x60" + "applicationPassword" + "\x60" + " for the configuration.\",\n      \"properties\": {\n        \"applicationName\": {\n          \"type\": \"string\"\n        },\n        \"applicationPassword\": {\n          \"type\": \"string\"\n        },\n        \"serverUrl\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the TestCrowdConfiguration tool (Status: 200, Content-Type: application/json)
const TestCrowdConfigurationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Test performed, results will be in the response message string.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **code** (Type: integer, int32):\n  - **message** (Type: string):\n"

// NewTestCrowdConfigurationMCPTool creates the MCP Tool instance for TestCrowdConfiguration
func NewTestCrowdConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"TestCrowdConfiguration",
		"Use this method to test a new or an existing Atlassian Crowd Server configuration.",
		[]byte(TestCrowdConfigurationInputSchema),
	)
}

// TestCrowdConfigurationHandler is the handler function for the TestCrowdConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func TestCrowdConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/config/crowd/test", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "TestCrowdConfiguration")
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
