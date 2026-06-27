package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the DeleteTag tool
const DeleteTagInputSchema = "{\n  \"properties\": {\n    \"organizationId\": {\n      \"description\": \"The organizationId assigned by IQ Server, corresponding to the application category tag you want to delete.\",\n      \"type\": \"string\"\n    },\n    \"tagId\": {\n      \"description\": \"The application category ID assigned by IQ Server, to be deleted.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"organizationId\",\n    \"tagId\"\n  ],\n  \"type\": \"object\"\n}"

// NewDeleteTagMCPTool creates the MCP Tool instance for DeleteTag
func NewDeleteTagMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DeleteTag",
		"Grouping applications with similar characteristics into categories makes policy management easier. You can then create a policy that applies to a specific category. Use this method to update an existing application category.Use this method to delete an existing application category.",
		[]byte(DeleteTagInputSchema),
	)
}

// DeleteTagHandler is the handler function for the DeleteTag tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DeleteTagHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/applicationCategories/organization/{organizationId}/{tagId}", args, []string{"organizationId", "tagId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DeleteTag")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
