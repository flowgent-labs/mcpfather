package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAllBoards tool
const GetAllBoardsInputSchema = "{\n  \"properties\": {\n    \"maxResults\": {\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"name\": {\n      \"type\": \"string\"\n    },\n    \"projectKeyOrId\": {\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"type\": {\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetAllBoards tool (Status: 200, Content-Type: application/json)
const GetAllBoardsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested boards, at the specified page of the results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/agile/1.0/board/10001'\n  - **type** (Type: string):\n      - Example: 'scrum'\n  - **id** (Type: integer, int64):\n      - Example: '10001'\n  - **name** (Type: string):\n      - Example: 'Scrum Board'\n"

// NewGetAllBoardsMCPTool creates the MCP Tool instance for GetAllBoards
func NewGetAllBoardsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllBoards",
		"Get all boards - Returns all boards. This only includes boards that the user has permission to view.",
		[]byte(GetAllBoardsInputSchema),
	)
}

// GetAllBoardsHandler is the handler function for the GetAllBoards tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllBoardsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAllBoards")
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
