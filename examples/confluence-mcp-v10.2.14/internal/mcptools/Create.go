package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Create tool
const CreateInputSchema = "{\n  \"properties\": {\n    \"body\": {}\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Create tool (Status: 200, Content-Type: application/json)
const CreateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns the new group if group is created successfully\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Create tool (Status: 400, Content-Type: application/json)
const CreateResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if request does not provide a name\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Create tool (Status: 403, Content-Type: application/json)
const CreateResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> returned if user does not have enough permission\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Create tool (Status: 409, Content-Type: application/json)
const CreateResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> returned if group with the same name already exists\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreateMCPTool creates the MCP Tool instance for Create
func NewCreateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Create",
		"Create group - Creates the given group identified by name.",
		[]byte(CreateInputSchema),
	)
}

// CreateHandler is the handler function for the Create tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/admin/group", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Create"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
