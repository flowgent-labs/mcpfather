package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostNewCodePeriodsUnset tool
const PostNewCodePeriodsUnsetInputSchema = "{\n  \"properties\": {\n    \"branch\": {\n      \"description\": \"Branch key\",\n      \"type\": \"string\"\n    },\n    \"project\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewPostNewCodePeriodsUnsetMCPTool creates the MCP Tool instance for PostNewCodePeriodsUnset
func NewPostNewCodePeriodsUnsetMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostNewCodePeriodsUnset",
		"Unsets the new code definition for a branch, project or global. - Unsets the new code definition for a branch, project or global. It requires the inherited New Code Definition to be compatible with the Clean as You Code methodology, and one of the following permissions: - 'Administer System' to change the global setting - 'Administer' rights for a specified component",
		[]byte(PostNewCodePeriodsUnsetInputSchema),
	)
}

// PostNewCodePeriodsUnsetHandler is the handler function for the PostNewCodePeriodsUnset tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostNewCodePeriodsUnsetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/new_code_periods/unset", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostNewCodePeriodsUnset")
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
