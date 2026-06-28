package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetCascadeStatus tool
const GetCascadeStatusInputSchema = "{\n  \"properties\": {\n    \"requestId\": {\n      \"description\": \"The cascade request ID to check status for\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"requestId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetCascadeStatus tool (Status: 200, Content-Type: application/json)
const GetCascadeStatusResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Cascade status retrieved successfully\n\n## Response Structure\n\n- Structure (Type: object):\n  - **status** (Type: string):\n      - Enum: ['PENDING', 'IN_PROGRESS', 'COMPLETED', 'NO_COMPONENTS_FOUND', 'FAILED']\n  - **evaluated** (Type: array):\n    - **Items** (Type: object):\n      - **componentId** (Type: string):\n      - **quarantined** (Type: boolean):\n      - **repositoryId** (Type: string):\n      - **repositoryManagerId** (Type: string):\n  - **failed** (Type: array):\n    - **[cyclic reference]**\n  - **pending** (Type: array):\n    - **[cyclic reference]**\n  - **referenceComponentHash** (Type: string):\n"

// NewGetCascadeStatusMCPTool creates the MCP Tool instance for GetCascadeStatus
func NewGetCascadeStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCascadeStatus",
		"Get the status of a cascade re-evaluation request.<p>Returns the current progress of a cascade re-evaluation operation including the list of components that have been evaluated and those still pending. The overall status will be 'pending' if any components are still being processed, or 'completed' if all components have been evaluated.<p>The response includes:<ul><li><b>status:</b> Overall status (PENDING, IN_PROGRESS, COMPLETED, NO_COMPONENTS_FOUND, FAILED)</li><li><b>referenceComponentHash:</b> The component hash that was re-evaluated</li><li><b>pending:</b> Components still being processed (PENDING status)</li><li><b>evaluated:</b> Components successfully re-evaluated (COMPLETED or NO_COMPONENTS_FOUND status)</li><li><b>failed:</b> Components that could not be re-evaluated (FAILED status)</li></ul><p>Permissions Required: Evaluate Components at Repository Managers level",
		[]byte(GetCascadeStatusInputSchema),
	)
}

// GetCascadeStatusHandler is the handler function for the GetCascadeStatus tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCascadeStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/repositories/cascade-reevaluate/status/{requestId}", args, []string{"requestId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetCascadeStatus")
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
