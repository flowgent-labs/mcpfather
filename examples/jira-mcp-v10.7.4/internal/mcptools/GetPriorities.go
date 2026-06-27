package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPriorities tool
const GetPrioritiesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetPriorities tool (Status: 200, Content-Type: application/json)
const GetPrioritiesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> List of priorities\n\n## Response Structure\n\n- Structure (Type: object):\n  - **iconUrl** (Type: string):\n      - Example: 'http://www.example.com/jira/images/icons/priorities/major.png'\n  - **id** (Type: string):\n      - Example: '1'\n  - **name** (Type: string):\n      - Example: 'Major'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/priority/1'\n  - **statusColor** (Type: string):\n      - Example: 'red'\n  - **description** (Type: string):\n      - Example: 'This is a description of the priority'\n"

// NewGetPrioritiesMCPTool creates the MCP Tool instance for GetPriorities
func NewGetPrioritiesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPriorities",
		"Get all issue priorities - Returns a list of all issue priorities",
		[]byte(GetPrioritiesInputSchema),
	)
}

// GetPrioritiesHandler is the handler function for the GetPriorities tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPrioritiesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/priority", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPriorities"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
