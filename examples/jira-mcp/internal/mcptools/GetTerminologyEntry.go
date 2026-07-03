package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetTerminologyEntry tool
const GetTerminologyEntryInputSchema = "{\n  \"properties\": {\n    \"originalName\": {\n      \"description\": \"A numeric StatusCategory id or a status category key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"originalName\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetTerminologyEntry tool (Status: 200, Content-Type: application/json)
const GetTerminologyEntryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns epic or sprint name.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **originalName** (Type: string):\n  - **originalNamePlural** (Type: string):\n  - **isDefault** (Type: boolean):\n  - **newName** (Type: string):\n  - **newNamePlural** (Type: string):\n"

// NewGetTerminologyEntryMCPTool creates the MCP Tool instance for GetTerminologyEntry
func NewGetTerminologyEntryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetTerminologyEntry",
		"Get epic or sprint name by original name - Returns epic or sprint name as specified in the {originalName} path param",
		[]byte(GetTerminologyEntryInputSchema),
	)
}

// GetTerminologyEntryHandler is the handler function for the GetTerminologyEntry tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetTerminologyEntryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/terminology/entries/{originalName}", args, []string{"originalName"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetTerminologyEntry")
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
