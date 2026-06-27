package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the SetProjectAssociationsForScheme tool
const SetProjectAssociationsForSchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Collection of projects, specified by id or key, to associate with this issue type scheme\",\n      \"properties\": {\n        \"idsOrKeys\": {\n          \"example\": [\n            \"100034\",\n            \"13543\",\n            \"FOOPROJ\",\n            \"BAZZPROJ\"\n          ],\n          \"items\": {\n            \"example\": \"[\\\"100034\\\",\\\"13543\\\",\\\"FOOPROJ\\\",\\\"BAZZPROJ\\\"]\",\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"schemeId\": {\n      \"description\": \"The id of the issue type scheme whose project associations we're replacing.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetProjectAssociationsForSchemeMCPTool creates the MCP Tool instance for SetProjectAssociationsForScheme
func NewSetProjectAssociationsForSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetProjectAssociationsForScheme",
		"Set project associations for scheme - Associates the given projects with the specified issue type scheme",
		[]byte(SetProjectAssociationsForSchemeInputSchema),
	)
}

// SetProjectAssociationsForSchemeHandler is the handler function for the SetProjectAssociationsForScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetProjectAssociationsForSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/issuetypescheme/{schemeId}/associations", args, []string{"schemeId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetProjectAssociationsForScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
