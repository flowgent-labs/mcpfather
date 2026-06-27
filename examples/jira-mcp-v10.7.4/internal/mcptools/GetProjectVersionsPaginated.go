package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetProjectVersionsPaginated tool
const GetProjectVersionsPaginatedInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Parameters to expand\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"How many results on the page should be included. Defaults to 50\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"orderBy\": {\n      \"description\": \"Ordering of the results\",\n      \"type\": \"string\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"Project id or project key\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"description\": \"The page offset, if not specified then defaults to 0\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetProjectVersionsPaginated tool (Status: 200, Content-Type: application/json)
const GetProjectVersionsPaginatedResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Project versions\n\n## Response Structure\n\n- Structure (Type: object):\n  - **startAt** (Type: integer, int64):\n  - **total** (Type: integer, int64):\n  - **values** (Type: array):\n    - **Items** (Type: object):\n  - **isLast** (Type: boolean):\n  - **maxResults** (Type: integer, int32):\n  - **nextPage** (Type: string, uri):\n  - **self** (Type: string, uri):\n"

// NewGetProjectVersionsPaginatedMCPTool creates the MCP Tool instance for GetProjectVersionsPaginated
func NewGetProjectVersionsPaginatedMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetProjectVersionsPaginated",
		"Get paginated project versions - Returns all versions for the specified project. Results are paginated. Results can be ordered by the following fields: sequence, name, startDate, releaseDate.",
		[]byte(GetProjectVersionsPaginatedInputSchema),
	)
}

// GetProjectVersionsPaginatedHandler is the handler function for the GetProjectVersionsPaginated tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetProjectVersionsPaginatedHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/project/{projectIdOrKey}/version", args, []string{"projectIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetProjectVersionsPaginated"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
