package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the VerifyConnectionAndGetApplications tool
const VerifyConnectionAndGetApplicationsInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the VerifyConnectionAndGetApplications tool (Status: 200, Content-Type: application/json)
const VerifyConnectionAndGetApplicationsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Connection verified successfully. Returns list of applications accessible to the user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationSummaries** (Type: array):\n    - **Items** (Type: object):\n      - **publicId** (Type: string):\n      - **id** (Type: string):\n      - **name** (Type: string):\n"

// NewVerifyConnectionAndGetApplicationsMCPTool creates the MCP Tool instance for VerifyConnectionAndGetApplications
func NewVerifyConnectionAndGetApplicationsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"VerifyConnectionAndGetApplications",
		"Use this method to verify that the authenticated user has required permissions for firewall operations and retrieve accessible applications.\n\nPermissions required: Evaluate Individual Components",
		[]byte(VerifyConnectionAndGetApplicationsInputSchema),
	)
}

// VerifyConnectionAndGetApplicationsHandler is the handler function for the VerifyConnectionAndGetApplications tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func VerifyConnectionAndGetApplicationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/connection/verify", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "VerifyConnectionAndGetApplications")
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
