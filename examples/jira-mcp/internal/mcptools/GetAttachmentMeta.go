package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAttachmentMeta tool
const GetAttachmentMetaInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAttachmentMeta tool (Status: 200, Content-Type: application/json)
const GetAttachmentMetaResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> JSON representation of the attachment capabilities. Consumers of this resource may also need to check if the logged in user has permission to upload or otherwise manipulate attachments using the com.atlassian.jira.rest.v2.permission.PermissionsResource\n\n## Response Structure\n\n- Structure (Type: object):\n  - **uploadLimit**: Upload limit in bytes (Type: integer, int64):\n      - Example: '1000000'\n  - **enabled** (Type: boolean):\n      - Example: 'true'\n"

// NewGetAttachmentMetaMCPTool creates the MCP Tool instance for GetAttachmentMeta
func NewGetAttachmentMetaMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAttachmentMeta",
		"Get attachment capabilities - Returns the meta information for an attachments, specifically if they are enabled and the maximum upload size allowed.",
		[]byte(GetAttachmentMetaInputSchema),
	)
}

// GetAttachmentMetaHandler is the handler function for the GetAttachmentMeta tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAttachmentMetaHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/attachment/meta", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAttachmentMeta")
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
