package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostPermissionsApplyTemplate tool
const PostPermissionsApplyTemplateInputSchema = "{\n  \"properties\": {\n    \"projectId\": {\n      \"description\": \"Project id\",\n      \"type\": \"string\"\n    },\n    \"projectKey\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    },\n    \"templateId\": {\n      \"description\": \"Template id\",\n      \"type\": \"string\"\n    },\n    \"templateName\": {\n      \"description\": \"Template name\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewPostPermissionsApplyTemplateMCPTool creates the MCP Tool instance for PostPermissionsApplyTemplate
func NewPostPermissionsApplyTemplateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostPermissionsApplyTemplate",
		"Apply a permission template to one project. - Apply a permission template to one project. The project id or project key must be provided. The template id or name must be provided. Requires the following permission: 'Administer System'.",
		[]byte(PostPermissionsApplyTemplateInputSchema),
	)
}

// PostPermissionsApplyTemplateHandler is the handler function for the PostPermissionsApplyTemplate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostPermissionsApplyTemplateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/permissions/apply_template", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostPermissionsApplyTemplate")
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
