package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration5 tool
const GetConfiguration5InputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration5 tool (Status: 200, Content-Type: application/json)
const GetConfiguration5ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "enabled" + "\x60" + " indicates if the configuration is enabled.</li><li>" + "\x60" + "usernameHeader" + "\x60" + " is the name of the HTTP request header field that contains the username. The default value is " + "\x60" + "REMOTE_USER" + "\x60" + ".</li><li>" + "\x60" + "csrfProtectionDisabled" + "\x60" + " indicates if Cross-Site Request Forgery (CSRF) protection is disabled. Used for backward compatibility with old client plugins.</li><li>" + "\x60" + "logoutUrl" + "\x60" + " is the redirect URL when a user logs out. If set to " + "\x60" + "null" + "\x60" + " the user will not be redirected.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **logoutUrl** (Type: string):\n  - **usernameHeader** (Type: string):\n  - **csrfProtectionDisabled** (Type: boolean):\n  - **enabled** (Type: boolean):\n"

// NewGetConfiguration5MCPTool creates the MCP Tool instance for GetConfiguration5
func NewGetConfiguration5MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration5",
		"Use this method to inspect an existing reverse proxy server configuration.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetConfiguration5InputSchema),
	)
}

// GetConfiguration5Handler is the handler function for the GetConfiguration5 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguration5Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/reverseProxyAuthentication", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration5")
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
