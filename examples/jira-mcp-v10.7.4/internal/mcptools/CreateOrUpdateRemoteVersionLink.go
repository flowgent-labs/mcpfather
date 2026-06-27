package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the CreateOrUpdateRemoteVersionLink tool
const CreateOrUpdateRemoteVersionLinkInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"JSON containing parameters to create the remote version link with\",\n      \"properties\": {\n        \"link\": {\n          \"example\": \"{\\\"rel\\\":\\\"issue\\\",\\\"url\\\":\\\"http://www.example.com/jira/rest/api/2/issue/10000\\\"}\",\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"example\": \"Issue 10000\",\n          \"type\": \"string\"\n        },\n        \"self\": {\n          \"example\": \"http://www.example.com/jira/rest/api/2/issue/10000\",\n          \"format\": \"uri\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"versionId\": {\n      \"description\": \"ID of the version.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"versionId\"\n  ],\n  \"type\": \"object\"\n}"

// NewCreateOrUpdateRemoteVersionLinkMCPTool creates the MCP Tool instance for CreateOrUpdateRemoteVersionLink
func NewCreateOrUpdateRemoteVersionLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateOrUpdateRemoteVersionLink",
		"Create or update remote version link without global ID - Create a remote version link via POST. The link's global ID will be taken from the JSON payload if provided; otherwise, it will be generated.",
		[]byte(CreateOrUpdateRemoteVersionLinkInputSchema),
	)
}

// CreateOrUpdateRemoteVersionLinkHandler is the handler function for the CreateOrUpdateRemoteVersionLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateOrUpdateRemoteVersionLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/version/{versionId}/remotelink", args, []string{"versionId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateOrUpdateRemoteVersionLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
