package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the RemoveProjectAssociation tool
const RemoveProjectAssociationInputSchema = "{\n  \"properties\": {\n    \"projIdOrKey\": {\n      \"description\": \"The id or key of the project that is to be un-associated with the issue type scheme\",\n      \"type\": \"string\"\n    },\n    \"schemeId\": {\n      \"description\": \"The id of the issue type scheme whose project association we're removing\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projIdOrKey\",\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewRemoveProjectAssociationMCPTool creates the MCP Tool instance for RemoveProjectAssociation
func NewRemoveProjectAssociationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RemoveProjectAssociation",
		"Remove given project association for specified scheme - For the specified issue type scheme, removes the given project association",
		[]byte(RemoveProjectAssociationInputSchema),
	)
}

// RemoveProjectAssociationHandler is the handler function for the RemoveProjectAssociation tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RemoveProjectAssociationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/rest/api/2/issuetypescheme/{schemeId}/associations/{projIdOrKey}", args, []string{"projIdOrKey", "schemeId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "RemoveProjectAssociation"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
