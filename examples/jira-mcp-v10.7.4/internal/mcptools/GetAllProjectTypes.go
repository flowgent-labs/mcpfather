package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAllProjectTypes tool
const GetAllProjectTypesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAllProjectTypes tool (Status: 200, Content-Type: application/json)
const GetAllProjectTypesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list with all the project types defined on the Jira instance\n\n## Response Structure\n\n- Structure (Type: object):\n  - **color** (Type: string):\n      - Example: '#FFFFFF'\n  - **descriptionI18nKey** (Type: string):\n      - Example: 'Project type for software projects'\n  - **formattedKey** (Type: string):\n      - Example: 'Software'\n  - **icon** (Type: string):\n      - Example: 'PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4NCjwhLS0gR2VuZXJhdG9yOiBBZG9iZSBJbGx1c3RyYXRvciAxOC4xLjEsIFNWRyBFeHBvcnQgUGx1Zy1JbiAuIFNWRyBWZXJzaW9uOiA2LjAwIEJ1aWxkIDApICAtLT4NCjxzdmcgdmVyc2lvbj0iMS4xIiBpZD0iTGF5ZXJfMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayIgeD0iMHB4IiB5PSIwcHgiDQoJIHZpZXdCb3g9IjAgMCAzMiAzMiIgZW5hYmxlLWJhY2tncm91bmQ9Im5ldyAwIDAgMzIgMzIiIHhtbDpzcGFjZT0icHJlc2VydmUiPg0KPGc+DQoJPHBhdGggZmlsbD0iIzY2NjY2NiIgZD0iTTE2LDBDNy4yLDAsMCw3LjIsMCwxNmMwLDguOCw3LjIsMTYsMTYsMTZjOC44LDAsMTYtNy4yLDE2LTE2QzMyLDcuMiwyNC44LDAsMTYsMHogTTI1LjcsMjMNCgkJYzAsMS44LTEuNCwzLjItMy4yLDMuMkg5LjJDNy41LDI2LjIsNiwyNC44LDYsMjNWOS44QzYsOCw3LjUsNi42LDkuMiw2LjZoMTMuMmMwLjIsMCwwLjQsMCwwLjcsMC4xbC0yLjgsMi44SDkuMg0KCQlDOSw5LjQsOC44LDkuNiw4LjgsOS44VjIzYzAsMC4yLDAuMiwwLjQsMC40LDAuNGgxMy4yYzAuMiwwLDAuNC0wLjIsMC40LTAuNHYtNS4zbDIuOC0yLjhWMjN6IE0xNS45LDIxLjNMMTEsMTYuNGwyLTJsMi45LDIuOQ0KCQlMMjYuNCw2LjhjMC42LDAuNywxLjIsMS41LDEuNywyLjNMMTUuOSwyMS4zeiIvPg0KPC9nPg0KPC9zdmc+'\n  - **key** (Type: string):\n      - Example: 'software'\n"

// NewGetAllProjectTypesMCPTool creates the MCP Tool instance for GetAllProjectTypes
func NewGetAllProjectTypesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllProjectTypes",
		"Get all project types - Returns all the project types defined on the Jira instance, not taking into account whether the license to use those project types is valid or not. In case of anonymous checks if they can access at least one project.",
		[]byte(GetAllProjectTypesInputSchema),
	)
}

// GetAllProjectTypesHandler is the handler function for the GetAllProjectTypes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllProjectTypesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/project/type", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAllProjectTypes"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
