package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the TestConfiguration tool
const TestConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Provide one or more values for the following in the JSON payload:\\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "hostname" + "\x60" + " - is the hostname or IP address of the SMTP server used for outgoing mail.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "port" + "\x60" + " - is the port number on which the SMTP server accepts email requests.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "password" + "\x60" + " - depends upon the value of " + "\x60" + "passwordIsIncluded" + "\x60" + ".\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "passwordIsIncluded" + "\x60" + " - if set to true, value must be provided for " + "\x60" + "password" + "\x60" + ", null is allowed.If set to false, the previous value will remain unchanged, provided that " + "\x60" + "hostname" + "\x60" + " and " + "\x60" + "port" + "\x60" + " are not changed.\\u003cli\\u003e" + "\x60" + "sslEnabled" + "\x60" + " - is a boolean flag indicating if the connection to the SMTP server should use SSL/TLSright from the start.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "startIsEnabled" + "\x60" + "- is a boolean flag indicating if the connection to the SMTP server should attempt toupgrade to SSL/TLS using the STARTTLS command.\\u003cli\\u003e" + "\x60" + "systemEmail" + "\x60" + " - is the email address used for the FROM header in emails sent by the IQ Server.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"hostname\": {\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"type\": \"string\"\n        },\n        \"passwordIsIncluded\": {\n          \"type\": \"boolean\"\n        },\n        \"port\": {\n          \"format\": \"int32\",\n          \"type\": \"integer\"\n        },\n        \"sslEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"startTlsEnabled\": {\n          \"type\": \"boolean\"\n        },\n        \"systemEmail\": {\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"recipientEmail\": {\n      \"description\": \"Enter the test recipient email address.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"recipientEmail\"\n  ],\n  \"type\": \"object\"\n}"

// NewTestConfigurationMCPTool creates the MCP Tool instance for TestConfiguration
func NewTestConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"TestConfiguration",
		"Use this method to test an SMTP configuration, without affecting the current configuration.\n\nPermissions required:Edit System Configuration and Users",
		[]byte(TestConfigurationInputSchema),
	)
}

// TestConfigurationHandler is the handler function for the TestConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func TestConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/config/mail/test/{recipientEmail}", args, []string{"recipientEmail"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "TestConfiguration")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
