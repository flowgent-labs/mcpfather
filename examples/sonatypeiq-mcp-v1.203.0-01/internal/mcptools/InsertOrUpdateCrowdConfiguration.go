package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the InsertOrUpdateCrowdConfiguration tool
const InsertOrUpdateCrowdConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON should include the " + "\x60" + "serverUrl" + "\x60" + ", " + "\x60" + "applicationName" + "\x60" + ", and the " + "\x60" + "applicationPassword" + "\x60" + " which will be used for authentication against the Atlassian Crowd Server.\\n\\nIf updating the " + "\x60" + "serverUrl" + "\x60" + ", the " + "\x60" + "applicationPassword" + "\x60" + " field is required.\",\n      \"properties\": {\n        \"applicationName\": {\n          \"type\": \"string\"\n        },\n        \"applicationPassword\": {\n          \"type\": \"string\"\n        },\n        \"serverUrl\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewInsertOrUpdateCrowdConfigurationMCPTool creates the MCP Tool instance for InsertOrUpdateCrowdConfiguration
func NewInsertOrUpdateCrowdConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"InsertOrUpdateCrowdConfiguration",
		"Use this method to create a new or update an existing Atlassian Crowd Server configuration.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(InsertOrUpdateCrowdConfigurationInputSchema),
	)
}

// InsertOrUpdateCrowdConfigurationHandler is the handler function for the InsertOrUpdateCrowdConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func InsertOrUpdateCrowdConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/crowd", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "InsertOrUpdateCrowdConfiguration")
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
