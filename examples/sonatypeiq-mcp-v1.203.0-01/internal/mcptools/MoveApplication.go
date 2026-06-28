package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the MoveApplication tool
const MoveApplicationInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the applicationId of the application to be moved.\",\n      \"type\": \"string\"\n    },\n    \"organizationId\": {\n      \"description\": \"Enter the organizationId of the destination organization.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the MoveApplication tool (Status: 200, Content-Type: application/json)
const MoveApplicationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Application moved successfully, with/without warnings. Warnings, if any, will appear in the response body.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **errors** (Type: array):\n    - **Items** (Type: string):\n  - **warnings** (Type: array):\n    - **Items** (Type: string):\n"

// NewMoveApplicationMCPTool creates the MCP Tool instance for MoveApplication
func NewMoveApplicationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"MoveApplication",
		"Use this method to move an application from one organization to another.\n\nPermissions required: Edit IQ Elements",
		[]byte(MoveApplicationInputSchema),
	)
}

// MoveApplicationHandler is the handler function for the MoveApplication tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveApplicationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/applications/{applicationId}/move/organization/{organizationId}", args, []string{"applicationId", "organizationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "MoveApplication")
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
