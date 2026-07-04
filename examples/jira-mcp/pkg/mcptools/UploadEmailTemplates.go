package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/pkg/helpers"
	"time"
)

// Input Schema for the UploadEmailTemplates tool
const UploadEmailTemplatesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewUploadEmailTemplatesMCPTool creates the MCP Tool instance for UploadEmailTemplates
func NewUploadEmailTemplatesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UploadEmailTemplates",
		"Update email templates with zip file - Extracts given zip file to temporary templates folder. If the folder already exists it will replace it's content",
		[]byte(UploadEmailTemplatesInputSchema),
	)
}

// UploadEmailTemplatesHandler is the handler function for the UploadEmailTemplates tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UploadEmailTemplatesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/zip"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/email-templates", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UploadEmailTemplates")
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
