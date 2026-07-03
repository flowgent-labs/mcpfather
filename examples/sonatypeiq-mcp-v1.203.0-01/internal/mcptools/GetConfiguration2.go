package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration2 tool
const GetConfiguration2InputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration2 tool (Status: 200, Content-Type: application/json)
const GetConfiguration2ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ol><li>" + "\x60" + "url" + "\x60" + " is the Jira server address.</li><li>" + "\x60" + "username" + "\x60" + " is the username used to connect to the Jira server.</li><li>" + "\x60" + "password" + "\x60" + " is the password used to authenticate on the Jira server.</li><li>" + "\x60" + "customFields" + "\x60" + " are any project issue type required fields defined in Jira.</li></ol>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **password** (Type: string):\n  - **url** (Type: string):\n  - **username** (Type: string):\n  - **customFields** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: object):\n"

// NewGetConfiguration2MCPTool creates the MCP Tool instance for GetConfiguration2
func NewGetConfiguration2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration2",
		"Use this method to retrieve an existing configuration for Jira.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetConfiguration2InputSchema),
	)
}

// GetConfiguration2Handler is the handler function for the GetConfiguration2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfiguration2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/jira", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration2")
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
