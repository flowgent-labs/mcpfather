package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the CreatePersonalSpace1 tool
const CreatePersonalSpace1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The personal space to be created\",\n      \"properties\": {\n        \"description\": {\n          \"type\": \"object\"\n        },\n        \"isPrivate\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"private\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CreatePersonalSpace1 tool (Status: 200, Content-Type: application/json)
const CreatePersonalSpace1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of a space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreatePersonalSpace1 tool (Status: 400, Content-Type: application/json)
const CreatePersonalSpace1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if there is invalid space description.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreatePersonalSpace1 tool (Status: 403, Content-Type: application/json)
const CreatePersonalSpace1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if current user does not have correct permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreatePersonalSpace1 tool (Status: 409, Content-Type: application/json)
const CreatePersonalSpace1ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> Returned if personal space already exists for user.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreatePersonalSpace1MCPTool creates the MCP Tool instance for CreatePersonalSpace1
func NewCreatePersonalSpace1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreatePersonalSpace1",
		"Creates the personal Space for self. - Creates a personal space for self.\n\nExample request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/api/space/personal"+"\x60"+"",
		[]byte(CreatePersonalSpace1InputSchema),
	)
}

// CreatePersonalSpace1Handler is the handler function for the CreatePersonalSpace1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreatePersonalSpace1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/space/personal", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreatePersonalSpace1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
