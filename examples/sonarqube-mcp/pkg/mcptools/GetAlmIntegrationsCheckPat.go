package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetAlmIntegrationsCheckPat tool
const GetAlmIntegrationsCheckPatInputSchema = "{\n  \"properties\": {\n    \"almSetting\": {\n      \"description\": \"DevOps Platform setting key\",\n      \"maxLength\": 200,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"almSetting\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAlmIntegrationsCheckPat tool (Status: 200, Content-Type: application/json)
const GetAlmIntegrationsCheckPatResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Structure (Type: object):\n"

// NewGetAlmIntegrationsCheckPatMCPTool creates the MCP Tool instance for GetAlmIntegrationsCheckPat
func NewGetAlmIntegrationsCheckPatMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAlmIntegrationsCheckPat",
		"Check validity of a Personal Access Token for the given DevOps Platform setting. - Check validity of a Personal Access Token for the given DevOps Platform setting Requires the 'Create Projects' permission",
		[]byte(GetAlmIntegrationsCheckPatInputSchema),
	)
}

// GetAlmIntegrationsCheckPatHandler is the handler function for the GetAlmIntegrationsCheckPat tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAlmIntegrationsCheckPatHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/alm_integrations/check_pat", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAlmIntegrationsCheckPat")
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
