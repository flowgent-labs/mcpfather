package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAllScreens tool
const GetAllScreensInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"type\": \"string\"\n    },\n    \"search\": {\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewGetAllScreensMCPTool creates the MCP Tool instance for GetAllScreens
func NewGetAllScreensMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllScreens",
		"Get available field screens - Adds field or custom field to the default tab.",
		[]byte(GetAllScreensInputSchema),
	)
}

// GetAllScreensHandler is the handler function for the GetAllScreens tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllScreensHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/screens", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAllScreens")
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
