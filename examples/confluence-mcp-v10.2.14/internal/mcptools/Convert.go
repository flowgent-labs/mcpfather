package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Convert tool
const ConvertInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"the body to convert from\"\n    },\n    \"expand\": {\n      \"description\": \"A comma separated list of properties to expand on the content. Default value: \\u003ccode\\u003ebody.storage,history,space,version,ancestors\\u003c/code\\u003e\",\n      \"type\": \"string\"\n    },\n    \"to\": {\n      \"description\": \"the representation to convert to.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"to\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Convert tool (Status: 200, Content-Type: application/json)
const ConvertResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns the converted body\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewConvertMCPTool creates the MCP Tool instance for Convert
func NewConvertMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Convert",
		"Convert body representation - Converts between content body representations. Not all representations can be converted to/from other formats. Supported conversions: \n\n- "+"\x60"+"storage -> view,export_view,styled_view,editor"+"\x60"+"\n- "+"\x60"+"editor -> storage"+"\x60"+"\n- "+"\x60"+"view -> None"+"\x60"+"\n- "+"\x60"+"export_view -> None"+"\x60"+"\n- "+"\x60"+"styled_view -> None"+"\x60"+"\n\nExample request URI(s):\n\n- "+"\x60"+"http://example.com/confluence/rest/api/contentbody/convert/view"+"\x60"+"",
		[]byte(ConvertInputSchema),
	)
}

// ConvertHandler is the handler function for the Convert tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ConvertHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/contentbody/convert/{to}", args, []string{"to"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Convert"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
