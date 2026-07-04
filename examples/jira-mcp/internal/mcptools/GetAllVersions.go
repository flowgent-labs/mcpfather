package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAllVersions tool
const GetAllVersionsInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"description\": \"The Id of the board that contains the requested versions.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of versions to return per page. Default: 50.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"released\": {\n      \"description\": \"Filters results to versions that are either released or unreleased. Valid values: true, false.\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"description\": \"The starting index of the returned versions. Base index: 0.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"boardId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAllVersions tool (Status: 200, Content-Type: application/json)
const GetAllVersionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the requested versions, at the specified page of the results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **projectId** (Type: integer, int64):\n      - Example: '10000'\n  - **released** (Type: boolean):\n      - Example: 'true'\n  - **overdue** (Type: boolean):\n      - Example: 'true'\n  - **self** (Type: string, uri):\n      - Example: 'http://localhost:8090/jira/rest/api/2/version/10000'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **releaseDateSet** (Type: boolean):\n      - Example: 'false'\n  - **startDateSet** (Type: boolean):\n      - Example: 'false'\n  - **userReleaseDate** (Type: string):\n      - Example: '2012-09-15T21:11:01.834+0000'\n  - **expand** (Type: string):\n      - Example: '10000'\n  - **project** (Type: string):\n      - Example: 'PXA'\n  - **moveUnfixedIssuesTo** (Type: string, uri):\n      - Example: 'http://localhost:8090/jira/rest/api/2/version/10000/move'\n  - **releaseDate** (Type: string, date-time):\n  - **archived** (Type: boolean):\n      - Example: 'false'\n  - **name** (Type: string):\n      - Example: 'New Version 1'\n  - **startDate** (Type: string, date-time):\n  - **description** (Type: string):\n      - Example: 'An excellent version'\n  - **userStartDate** (Type: string):\n      - Example: '2012-08-15T21:11:01.834+0000'\n"

// NewGetAllVersionsMCPTool creates the MCP Tool instance for GetAllVersions
func NewGetAllVersionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllVersions",
		"Get all versions from a board - Returns all versions from a board, for a given board Id. This only includes versions that the user has permission to view. Note, if the user does not have permission to view the board, no versions will be returned at all. Returned versions are ordered by the name of the project from which they belong and then by sequence defined by user.",
		[]byte(GetAllVersionsInputSchema),
	)
}

// GetAllVersionsHandler is the handler function for the GetAllVersions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllVersionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/agile/1.0/board/{boardId}/version", args, []string{"boardId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAllVersions")
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
