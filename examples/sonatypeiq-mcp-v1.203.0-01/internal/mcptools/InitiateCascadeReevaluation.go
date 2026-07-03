package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the InitiateCascadeReevaluation tool
const InitiateCascadeReevaluationInputSchema = "{\n  \"properties\": {\n    \"componentHash\": {\n      \"description\": \"The component hash to re-evaluate across all repositories\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"componentHash\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the InitiateCascadeReevaluation tool (Status: 200, Content-Type: application/json)
const InitiateCascadeReevaluationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Cascade re-evaluation initiated successfully. The response contains statusUrl with a requestId, which can be used to check the cascade re-evaluation status using the GET method.A requestId for a cascade re-evaluation only lasts 24 hours before being deleted.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **statusUrl** (Type: string):\n"

// NewInitiateCascadeReevaluationMCPTool creates the MCP Tool instance for InitiateCascadeReevaluation
func NewInitiateCascadeReevaluationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"InitiateCascadeReevaluation",
		"Initiate cascade re-evaluation for a component across repository hierarchies.<p>This operation asynchronously re-evaluates the specified component across all repositories where the component exists.<p>The system will automatically discover all eligible repositories based on component presence.<p>Permissions Required: Evaluate Components at Repository Managers level",
		[]byte(InitiateCascadeReevaluationInputSchema),
	)
}

// InitiateCascadeReevaluationHandler is the handler function for the InitiateCascadeReevaluation tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func InitiateCascadeReevaluationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/firewall/repositories/cascade-reevaluate/componentHash/{componentHash}", args, []string{"componentHash"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "InitiateCascadeReevaluation")
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
