package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the ValidateSourceControlConfig tool
const ValidateSourceControlConfigInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the applicationId for which you want to validate the composite source control configuration.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ValidateSourceControlConfig tool (Status: 200, Content-Type: application/json)
const ValidateSourceControlConfigResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response shows if the composite source control configuration for the application is valid.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **repoPrivate** (Type: object):\n    - **valid** (Type: boolean):\n    - **message** (Type: string):\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n  - **[cyclic reference]**\n"

// NewValidateSourceControlConfigMCPTool creates the MCP Tool instance for ValidateSourceControlConfig
func NewValidateSourceControlConfigMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ValidateSourceControlConfig",
		"Use this method to validate the composite source control configuration.\n\nPermissions required: View IQ Elements",
		[]byte(ValidateSourceControlConfigInputSchema),
	)
}

// ValidateSourceControlConfigHandler is the handler function for the ValidateSourceControlConfig tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ValidateSourceControlConfigHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/compositeSourceControlConfigValidator/application/{applicationId}", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ValidateSourceControlConfig")
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
