package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAllTerminologyEntries tool
const GetAllTerminologyEntriesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAllTerminologyEntries tool (Status: 200, Content-Type: application/json)
const GetAllTerminologyEntriesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all defined names for the default words 'epic' and 'sprint'\n\n## Response Structure\n\n- Structure (Type: object):\n  - **originalName** (Type: string):\n  - **originalNamePlural** (Type: string):\n  - **isDefault** (Type: boolean):\n  - **newName** (Type: string):\n  - **newNamePlural** (Type: string):\n"

// NewGetAllTerminologyEntriesMCPTool creates the MCP Tool instance for GetAllTerminologyEntries
func NewGetAllTerminologyEntriesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllTerminologyEntries",
		"Get all defined names for 'epic' and 'sprint' - Returns a list of all defined names for the default words 'epic' and 'sprint'",
		[]byte(GetAllTerminologyEntriesInputSchema),
	)
}

// GetAllTerminologyEntriesHandler is the handler function for the GetAllTerminologyEntries tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllTerminologyEntriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/terminology/entries", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAllTerminologyEntries"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
