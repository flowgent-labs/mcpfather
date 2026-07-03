package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetReindexProgress tool
const GetReindexProgressInputSchema = "{\n  \"properties\": {\n    \"taskId\": {\n      \"description\": \"The id of an indexing task you wish to obtain details on. If omitted, then defaults to the standard behaviour and returns information on the active reindex task, or the last task to run if no reindex is taking place.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetReindexProgress tool (Status: 200, Content-Type: application/json)
const GetReindexProgressResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a representation of the progress of the re-index operation.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **type** (Type: string):\n      - Example: 'FOREGROUND'\n      - Enum: ['FOREGROUND', 'BACKGROUND', 'BACKGROUND_PREFFERED', 'BACKGROUND_PREFERRED']\n  - **currentProgress** (Type: integer, int64):\n      - Example: '0'\n  - **currentSubTask** (Type: string):\n      - Example: 'Currently reindexing Change History'\n  - **finishTime** (Type: string, date-time):\n  - **progressUrl** (Type: string):\n      - Example: 'http://localhost:8080/jira'\n  - **startTime** (Type: string, date-time):\n  - **submittedTime** (Type: string, date-time):\n  - **success** (Type: boolean):\n      - Example: 'true'\n"

// NewGetReindexProgressMCPTool creates the MCP Tool instance for GetReindexProgress
func NewGetReindexProgressMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetReindexProgress",
		"Get reindex progress - Returns information on the system reindexes. If a reindex is currently taking place then information about this reindex is returned. If there is no active index task, then returns information about the latest reindex task run, otherwise returns a 404 indicating that no reindex has taken place.",
		[]byte(GetReindexProgressInputSchema),
	)
}

// GetReindexProgressHandler is the handler function for the GetReindexProgress tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetReindexProgressHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/reindex/progress", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetReindexProgress")
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
