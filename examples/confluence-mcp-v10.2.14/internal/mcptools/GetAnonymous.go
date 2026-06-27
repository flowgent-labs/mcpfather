package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetAnonymous tool
const GetAnonymousInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"properties to expand on the user.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetAnonymous tool (Status: 200, Content-Type: application/json)
const GetAnonymousResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of a user.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetAnonymous tool (Status: 403, Content-Type: application/json)
const GetAnonymousResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling user does not have permission to use confluence.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetAnonymousMCPTool creates the MCP Tool instance for GetAnonymous
func NewGetAnonymousMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAnonymous",
		"Get information about anonymous user type - Get information about how anonymous is represented in Confluence. Example request URI(s):\n\n"+"\x60"+"http://example.com/confluence/rest/api/user/anonymous"+"\x60"+"",
		[]byte(GetAnonymousInputSchema),
	)
}

// GetAnonymousHandler is the handler function for the GetAnonymous tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAnonymousHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/user/anonymous", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAnonymous"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
