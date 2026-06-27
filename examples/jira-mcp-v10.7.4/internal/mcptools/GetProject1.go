package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProject1 tool
const GetProject1InputSchema = "{\n  \"properties\": {\n    \"key\": {\n      \"description\": \"The project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetProject1 tool (Status: 200, Content-Type: application/json)
const GetProject1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns an ErrorCollection containing any validation errors for the project key. If the project key is valid, the ErrorCollection will be empty.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **errorMessages** (Type: array):\n    - **Items** (Type: string):\n  - **errors** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: string):\n"

// NewGetProject1MCPTool creates the MCP Tool instance for GetProject1
func NewGetProject1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProject1",
		"Get project key validation - Validates a project key.",
		[]byte(GetProject1InputSchema),
	)
}

// GetProject1Handler is the handler function for the GetProject1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProject1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/projectvalidate/key", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProject1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
