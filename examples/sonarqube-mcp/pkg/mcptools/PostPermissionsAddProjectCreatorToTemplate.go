package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostPermissionsAddProjectCreatorToTemplate tool
const PostPermissionsAddProjectCreatorToTemplateInputSchema = "{\n  \"properties\": {\n    \"permission\": {\n      \"description\": \"Permission - Possible values for project permissions admin, codeviewer, issueadmin, securityhotspotadmin, architectureadmin, scan, user\",\n      \"enum\": [\n        \"admin\",\n        \"codeviewer\",\n        \"issueadmin\",\n        \"securityhotspotadmin\",\n        \"architectureadmin\",\n        \"scan\",\n        \"user\"\n      ],\n      \"type\": \"string\"\n    },\n    \"templateId\": {\n      \"description\": \"Template id\",\n      \"type\": \"string\"\n    },\n    \"templateName\": {\n      \"description\": \"Template name\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"permission\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostPermissionsAddProjectCreatorToTemplateMCPTool creates the MCP Tool instance for PostPermissionsAddProjectCreatorToTemplate
func NewPostPermissionsAddProjectCreatorToTemplateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostPermissionsAddProjectCreatorToTemplate",
		"Add a project creator to a permission template. - Add a project creator to a permission template. Requires the following permission: 'Administer System'.",
		[]byte(PostPermissionsAddProjectCreatorToTemplateInputSchema),
	)
}

// PostPermissionsAddProjectCreatorToTemplateHandler is the handler function for the PostPermissionsAddProjectCreatorToTemplate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostPermissionsAddProjectCreatorToTemplateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/permissions/add_project_creator_to_template", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostPermissionsAddProjectCreatorToTemplate")
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
