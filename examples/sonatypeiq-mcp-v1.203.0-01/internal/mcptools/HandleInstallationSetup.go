package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the HandleInstallationSetup tool
const HandleInstallationSetupInputSchema = "{\n  \"properties\": {\n    \"code\": {\n      \"description\": \"OAuth authorization code\",\n      \"minLength\": 1,\n      \"type\": \"string\"\n    },\n    \"installation_id\": {\n      \"description\": \"GitHub App installation ID\",\n      \"format\": \"int64\",\n      \"minimum\": 1,\n      \"type\": \"integer\"\n    },\n    \"state\": {\n      \"description\": \"State token for CSRF protection\",\n      \"minLength\": 1,\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"code\",\n    \"installation_id\",\n    \"state\"\n  ],\n  \"type\": \"object\"\n}"

// NewHandleInstallationSetupMCPTool creates the MCP Tool instance for HandleInstallationSetup
func NewHandleInstallationSetupMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"HandleInstallationSetup",
		"Handle GitHub App installation setup callback with OAuth + PKCE - Process the redirect from GitHub after OAuth authorization, validate state token, exchange OAuth code with PKCE verification, verify user ownership, configure the installation for the specified organization/application, and redirect to the configuration page",
		[]byte(HandleInstallationSetupInputSchema),
	)
}

// HandleInstallationSetupHandler is the handler function for the HandleInstallationSetup tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func HandleInstallationSetupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/githubApp/setupInstallation", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "HandleInstallationSetup")
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
