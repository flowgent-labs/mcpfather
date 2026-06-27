package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the Reindex tool
const ReindexInputSchema = "{\n  \"properties\": {\n    \"indexChangeHistory\": {\n      \"default\": false,\n      \"description\": \"Indicates that changeHistory should also be reindexed. Not relevant for foreground reindex, where changeHistory is always reindexed.\",\n      \"type\": \"boolean\"\n    },\n    \"indexComments\": {\n      \"default\": false,\n      \"description\": \"Indicates that comments should also be reindexed. Not relevant for foreground reindex, where comments are always reindexed.\",\n      \"type\": \"boolean\"\n    },\n    \"indexWorklogs\": {\n      \"default\": false,\n      \"description\": \"Indicates that worklogs should also be reindexed. Not relevant for foreground reindex, where worklogs are always reindexed.\",\n      \"type\": \"boolean\"\n    },\n    \"type\": {\n      \"description\": \"Case insensitive String indicating type of reindex. If omitted, then defaults to BACKGROUND_PREFERRED.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Reindex tool (Status: 202, Content-Type: application/json)
const ReindexResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 202\n\n**Content-Type:** application/json\n\n> Returns a representation of the progress of the re-index operation.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **currentProgress** (Type: integer, int64):\n      - Example: '0'\n  - **currentSubTask** (Type: string):\n      - Example: 'Currently reindexing Change History'\n  - **finishTime** (Type: string, date-time):\n  - **progressUrl** (Type: string):\n      - Example: 'http://localhost:8080/jira'\n  - **startTime** (Type: string, date-time):\n  - **submittedTime** (Type: string, date-time):\n  - **success** (Type: boolean):\n      - Example: 'true'\n  - **type** (Type: string):\n      - Example: 'FOREGROUND'\n      - Enum: ['FOREGROUND', 'BACKGROUND', 'BACKGROUND_PREFFERED', 'BACKGROUND_PREFERRED']\n"

// NewReindexMCPTool creates the MCP Tool instance for Reindex
func NewReindexMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Reindex",
		"Start a reindex operation - Kicks off a reindex. Need Admin permissions to perform this reindex.",
		[]byte(ReindexInputSchema),
	)
}

// ReindexHandler is the handler function for the Reindex tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ReindexHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/rest/api/2/reindex", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Reindex"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
