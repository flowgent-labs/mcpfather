package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueLinkTypes tool
const GetIssueLinkTypesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueLinkTypes tool (Status: 200, Content-Type: application/json)
const GetIssueLinkTypesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all available issue link types.\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewGetIssueLinkTypesMCPTool creates the MCP Tool instance for GetIssueLinkTypes
func NewGetIssueLinkTypesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueLinkTypes",
		"Get list of available issue link types - Returns a list of available issue link types, if issue linking is enabled.",
		[]byte(GetIssueLinkTypesInputSchema),
	)
}

// GetIssueLinkTypesHandler is the handler function for the GetIssueLinkTypes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueLinkTypesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issueLinkType", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueLinkTypes"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
