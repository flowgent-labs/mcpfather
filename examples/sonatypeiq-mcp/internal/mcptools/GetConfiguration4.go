package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration4 tool
const GetConfiguration4InputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration4 tool (Status: 200, Content-Type: application/json)
const GetConfiguration4ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "hostname" + "\x60" + " is host name or IP address of the HTTP proxy server to use for outgoing HTTP connections.</li><li>" + "\x60" + "port" + "\x60" + " is the port number for the HTTP proxy server.</li><li>" + "\x60" + "username" + "\x60" + " is the username needed to authenticate with the HTTP proxy server.</li><li>" + "\x60" + "password" + "\x60" + " is always null, never included for security purposes.</li><li>" + "\x60" + "passwordIsIncluded" + "\x60" + " is always FALSE </li><li>" + "\x60" + "excludeHosts" + "\x60" + " is a list of host names that are to be excluded from using the HTTP proxy server.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **port** (Type: integer, int32):\n  - **username** (Type: string):\n  - **excludeHosts** (Type: array):\n    - **Items** (Type: string):\n  - **hostname** (Type: string):\n  - **password** (Type: string):\n  - **passwordIsIncluded** (Type: boolean):\n"

// NewGetConfiguration4MCPTool creates the MCP Tool instance for GetConfiguration4
func NewGetConfiguration4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration4",
		"Use this method to inspect an existing HTTP proxy server configuration.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetConfiguration4InputSchema),
	)
}

// GetConfiguration4Handler is the handler function for the GetConfiguration4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguration4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/httpProxyServer", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration4")
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
