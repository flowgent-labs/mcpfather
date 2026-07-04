package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Spaces tool
const SpacesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"type\": \"string\"\n    },\n    \"contentLabel\": {\n      \"description\": \"filter the list of spaces returned by content containing provided label.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the spaces.\",\n      \"type\": \"string\"\n    },\n    \"favourite\": {\n      \"description\": \"filter the list of spaces returned by favourites.\",\n      \"type\": \"boolean\"\n    },\n    \"hasRetentionPolicy\": {\n      \"description\": \"filter the list of spaces returned by retention policy.\",\n      \"type\": \"boolean\"\n    },\n    \"label\": {\n      \"description\": \"filter the list of spaces returned by label.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"the limit of the number of spaces to return, this may be restricted by fixed system limits\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"spaceId\": {\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"spaceIds\": {\n      \"description\": \"the ids of the spaces to fetch information from. Cannot be used in conjunction with spaceKey(s)\",\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the key of the space to fetch information from.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\"\n    },\n    \"spaceKeySingle\": {\n      \"type\": \"string\"\n    },\n    \"spaceKeys\": {\n      \"description\": \"the keys of the spaces to fetch information from.\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"the start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"status\": {\n      \"description\": \"filter the list of spaces returned by status (current, archived).\",\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"description\": \"filter the list of spaces returned by type (global, personal).\",\n      \"type\": \"string\"\n    },\n    \"xoauth_requestor_id\": {\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Spaces tool (Status: 200, Content-Type: application/json)
const SpacesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns an array of full JSON representations of found space.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n  - **limit** (Type: number):\n      - Example: '25'\n"

// NewSpacesMCPTool creates the MCP Tool instance for Spaces
func NewSpacesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Spaces",
		"Get spaces by key - Returns information about a number of spaces. \n\nExample request URI(s): \n\n"+"\x60"+"http://example.com/confluence/rest/api/space?spaceKey=TST&spaceKey=ds"+"\x60"+"",
		[]byte(SpacesInputSchema),
	)
}

// SpacesHandler is the handler function for the Spaces tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SpacesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Spaces")
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
