package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration1 tool
const GetConfiguration1InputSchema = "{\n  \"properties\": {\n    \"property\": {\n      \"description\": \"Enter the names of the system properties. Values provided for name are case-sensitive.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration1 tool (Status: 200, Content-Type: application/json)
const GetConfiguration1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains all the requested properties and the corresponding values.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **Additional Properties**:\n    - **property value** (Type: object):\n"

// NewGetConfiguration1MCPTool creates the MCP Tool instance for GetConfiguration1
func NewGetConfiguration1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration1",
		"Use this method to retrieve the configured value for an IQ Server system property.\n\nPermissions required: Edit System Configuration and Users or system property dependent",
		[]byte(GetConfiguration1InputSchema),
	)
}

// GetConfiguration1Handler is the handler function for the GetConfiguration1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguration1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration1")
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
