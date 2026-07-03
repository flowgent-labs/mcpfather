package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration3 tool
const GetConfiguration3InputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration3 tool (Status: 200, Content-Type: application/json)
const GetConfiguration3ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "hostname" + "\x60" + " is the hostname or IP address of the SMTP server used for outgoing mail.</li><li>" + "\x60" + "port" + "\x60" + " is the port number on which the SMTP server accepts email requests.</li><li>" + "\x60" + "username" + "\x60" + " is the username to authenticate users on the SMTP server.</li><li>" + "\x60" + "password" + "\x60" + " is always null, never included for security purposes for this method.</Li><li>" + "\x60" + "passwordIsIncluded" + "\x60" + " is always FALSE for this method.</li><li>" + "\x60" + "sslEnabled" + "\x60" + " is a boolean flag indicating if the connection to the SMTP server should use SSL/TLS right from the start.</li><li>" + "\x60" + "startIsEnabled" + "\x60" + " is a boolean flag indicating if the connection to the SMTP server should attempt to upgrade to SSL/TLS using the STARTTLS command.</li><li>" + "\x60" + "systemEmail" + "\x60" + " is the email address used for the FROM header in emails sent by the IQ Server.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **startTlsEnabled** (Type: boolean):\n  - **systemEmail** (Type: string):\n  - **username** (Type: string):\n  - **hostname** (Type: string):\n  - **password** (Type: string):\n  - **passwordIsIncluded** (Type: boolean):\n  - **port** (Type: integer, int32):\n  - **sslEnabled** (Type: boolean):\n"

// NewGetConfiguration3MCPTool creates the MCP Tool instance for GetConfiguration3
func NewGetConfiguration3MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration3",
		"Use this method to review the configuration for an SMTP server.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetConfiguration3InputSchema),
	)
}

// GetConfiguration3Handler is the handler function for the GetConfiguration3 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguration3Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/mail", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration3")
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
