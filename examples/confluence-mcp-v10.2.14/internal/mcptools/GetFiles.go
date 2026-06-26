package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetFiles tool
const GetFilesInputSchema = "{\n  \"properties\": {\n    \"jobScope\": {\n      \"description\": \"name of type of restore job (SITE or SPACE or null), if null, all backup files are listed\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetFiles tool (Status: 200, Content-Type: application/json)
const GetFilesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of FileInfo objects, containing fileName, fileCreationTime, fileSize, and jobScope.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **name** (Type: string):\n        - Example: 'backup-2020-01-01-01-01-01.zip'\n    - **size** (Type: integer, int64):\n        - Example: '1000'\n    - **creationTime** (Type: string):\n        - Example: '2020-01-01T01:01:01.000Z'\n    - **jobScope** (Type: string):\n        - Example: 'SITE'\n        - Enum: ['SPACE', 'SITE']\n"

// Response Template for the GetFiles tool (Status: 400, Content-Type: application/json)
const GetFilesResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if user is not a system administrator\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetFilesMCPTool creates the MCP Tool instance for GetFiles
func NewGetFilesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetFiles",
		"Get files in restore directory - returns list of information on files in conf-home/restore/(jobScope).",
		[]byte(GetFilesInputSchema),
	)
}

// GetFilesHandler is the handler function for the GetFiles tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/backup-restore/restore/files", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetFiles"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
