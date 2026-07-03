package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetConfiguration4 tool
const SetConfiguration4InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON could include: \\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "hostname" + "\x60" + " is host name or IP address of the HTTP proxy server to use for outgoing HTTP connections.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "port" + "\x60" + " is the port number for the HTTP proxy server.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "username" + "\x60" + " is the username used to authenticate with the HTTP proxy server.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "password" + "\x60" + " is the password used for authentication with the HTTP proxy server.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "passwordIsIncluded" + "\x60" + " should be " + "\x60" + "true" + "\x60" + " if password is included in the request.\\u003cul\\u003e\\u003cli\\u003eIf " + "\x60" + "true" + "\x60" + " but the password is not included the password will be considered as " + "\x60" + "null" + "\x60" + ".\\u003c/li\\u003e\\u003cli\\u003eCan be " + "\x60" + "false" + "\x60" + " for update operations that do not a require password change. Note that updating the hostname and port requires a password to be provided.\\u003c/li\\u003e \\u003c/ul\\u003e\\u003cli\\u003e" + "\x60" + "excludeHosts" + "\x60" + " is a list of host names that are to be excluded from using the HTTP proxy server.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"excludeHosts\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"hostname\": {\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"type\": \"string\"\n        },\n        \"passwordIsIncluded\": {\n          \"type\": \"boolean\"\n        },\n        \"port\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewSetConfiguration4MCPTool creates the MCP Tool instance for SetConfiguration4
func NewSetConfiguration4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetConfiguration4",
		"Use this method to create or update an existing HTTP proxy server configuration.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(SetConfiguration4InputSchema),
	)
}

// SetConfiguration4Handler is the handler function for the SetConfiguration4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetConfiguration4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/httpProxyServer", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetConfiguration4")
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
