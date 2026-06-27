package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetVersion tool
const GetVersionInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"ID of the version.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetVersion tool (Status: 200, Content-Type: application/json)
const GetVersionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the version was found.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string, uri):\n      - Example: 'http://localhost:8090/jira/rest/api/2/version/10000'\n  - **startDate** (Type: string, date-time):\n  - **projectId** (Type: integer, int64):\n      - Example: '10000'\n  - **description** (Type: string):\n      - Example: 'An excellent version'\n  - **expand** (Type: string):\n      - Example: '10000'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **overdue** (Type: boolean):\n      - Example: 'true'\n  - **releaseDate** (Type: string, date-time):\n  - **project** (Type: string):\n      - Example: 'PXA'\n  - **archived** (Type: boolean):\n      - Example: 'false'\n  - **released** (Type: boolean):\n      - Example: 'true'\n  - **userStartDate** (Type: string):\n      - Example: '2012-08-15T21:11:01.834+0000'\n  - **name** (Type: string):\n      - Example: 'New Version 1'\n  - **userReleaseDate** (Type: string):\n      - Example: '2012-09-15T21:11:01.834+0000'\n  - **moveUnfixedIssuesTo** (Type: string, uri):\n      - Example: 'http://localhost:8090/jira/rest/api/2/version/10000/move'\n  - **startDateSet** (Type: boolean):\n      - Example: 'false'\n  - **releaseDateSet** (Type: boolean):\n      - Example: 'false'\n"

// NewGetVersionMCPTool creates the MCP Tool instance for GetVersion
func NewGetVersionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetVersion",
		"Get version details - Returns a version.",
		[]byte(GetVersionInputSchema),
	)
}

// GetVersionHandler is the handler function for the GetVersion tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetVersionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/version/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetVersion"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
