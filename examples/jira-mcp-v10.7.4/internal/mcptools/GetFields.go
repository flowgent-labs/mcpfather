package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetFields tool
const GetFieldsInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetFields tool (Status: 200, Content-Type: application/json)
const GetFieldsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all fields\n\n## Response Structure\n\n- Structure (Type: object):\n  - **navigable** (Type: boolean):\n      - Example: 'true'\n  - **orderable** (Type: boolean):\n      - Example: 'true'\n  - **schema** (Type: object):\n      - Example: '{}'\n    - **items** (Type: string):\n        - Example: 'null'\n    - **system** (Type: string):\n        - Example: 'summary'\n    - **type** (Type: string):\n        - Example: 'string'\n    - **custom** (Type: string):\n        - Example: 'null'\n    - **customId** (Type: integer, int64):\n  - **searchable** (Type: boolean):\n      - Example: 'true'\n  - **clauseNames** (Type: array):\n      - Unique Items: true\n      - Example: '\"[description]\"'\n    - **Items** (Type: string):\n        - Example: '[description]'\n  - **custom** (Type: boolean):\n      - Example: 'false'\n  - **id** (Type: string):\n      - Example: 'description'\n  - **name** (Type: string):\n      - Example: 'Description'\n"

// NewGetFieldsMCPTool creates the MCP Tool instance for GetFields
func NewGetFieldsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetFields",
		"Get all fields, both System and Custom - Returns a list of all fields, both System and Custom",
		[]byte(GetFieldsInputSchema),
	)
}

// GetFieldsHandler is the handler function for the GetFields tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetFieldsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/field", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetFields"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
