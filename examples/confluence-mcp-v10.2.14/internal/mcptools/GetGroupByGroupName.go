package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetGroupByGroupName tool
const GetGroupByGroupNameInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"type\": \"string\"\n    },\n    \"groupName\": {\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetGroupByGroupName tool (Status: 200, Content-Type: application/json)
const GetGroupByGroupNameResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The user group with the group name\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetGroupByGroupName tool (Status: 403, Content-Type: application/json)
const GetGroupByGroupNameResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> The calling user does not have permission to view groups\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetGroupByGroupNameMCPTool creates the MCP Tool instance for GetGroupByGroupName
func NewGetGroupByGroupNameMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetGroupByGroupName",
		"Get group by name - Get the user group with the group name",
		[]byte(GetGroupByGroupNameInputSchema),
	)
}

// GetGroupByGroupNameHandler is the handler function for the GetGroupByGroupName tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetGroupByGroupNameHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/group/info", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetGroupByGroupName"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
