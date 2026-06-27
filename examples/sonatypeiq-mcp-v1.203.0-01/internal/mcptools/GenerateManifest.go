package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GenerateManifest tool
const GenerateManifestInputSchema = "{\n  \"properties\": {\n    \"organizationName\": {\n      \"description\": \"GitHub organization name\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Owner (organization/application) ID\",\n      \"minLength\": 1,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGenerateManifestMCPTool creates the MCP Tool instance for GenerateManifest
func NewGenerateManifestMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GenerateManifest",
		"Generate GitHub App manifest - Generate a GitHub App manifest for registration. Returns manifest JSON with a state token for CSRF protection. The state token is cryptographically secure, single-use, and expires after 10 minutes. Submit the manifest to GitHub's app creation flow, which will redirect back to IQ Server with the state token for validation. \n\n**Permissions Required:** Configure System Configuration and Users",
		[]byte(GenerateManifestInputSchema),
	)
}

// GenerateManifestHandler is the handler function for the GenerateManifest tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GenerateManifestHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/githubApp/manifest", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GenerateManifest")
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
