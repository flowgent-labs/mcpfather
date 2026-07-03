package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetConfiguration5 tool
const SetConfiguration5InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON could include: \\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "enabled" + "\x60" + " indicates if the configuration is enabled.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "usernameHeader" + "\x60" + " is the name of the HTTP request header field that contains the username. The default value is " + "\x60" + "REMOTE_USER" + "\x60" + ".\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "csrfProtectionDisabled" + "\x60" + " indicates if Cross-Site Request Forgery (CSRF) protection is disabled. Used for backward compatibility with old client plugins.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "logoutUrl" + "\x60" + " is the redirect URL when a user logs out. If set to " + "\x60" + "null" + "\x60" + " the user will not be redirected.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"csrfProtectionDisabled\": {\n          \"type\": \"boolean\"\n        },\n        \"enabled\": {\n          \"type\": \"boolean\"\n        },\n        \"logoutUrl\": {\n          \"type\": \"string\"\n        },\n        \"usernameHeader\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewSetConfiguration5MCPTool creates the MCP Tool instance for SetConfiguration5
func NewSetConfiguration5MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetConfiguration5",
		"Use this method to configure the reverse proxy server.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(SetConfiguration5InputSchema),
	)
}

// SetConfiguration5Handler is the handler function for the SetConfiguration5 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetConfiguration5Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/reverseProxyAuthentication", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetConfiguration5")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
