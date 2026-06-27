package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetDefaultColorScheme tool
const GetDefaultColorSchemeInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetDefaultColorScheme tool (Status: 200, Content-Type: application/json)
const GetDefaultColorSchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of color scheme including light and dark theme\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetDefaultColorSchemeMCPTool creates the MCP Tool instance for GetDefaultColorScheme
func NewGetDefaultColorSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetDefaultColorScheme",
		"Get default global color scheme - Get default color scheme for the instance\n\n",
		[]byte(GetDefaultColorSchemeInputSchema),
	)
}

// GetDefaultColorSchemeHandler is the handler function for the GetDefaultColorScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetDefaultColorSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/color-scheme/default", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetDefaultColorScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
