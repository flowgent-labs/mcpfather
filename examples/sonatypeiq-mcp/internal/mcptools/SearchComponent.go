package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the SearchComponent tool
const SearchComponentInputSchema = "{\n  \"properties\": {\n    \"componentIdentifier\": {\n      \"description\": \"Specify the componentIdentifier object containing the format and coordinates.\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"hash\": {\n      \"description\": \"Enter the component hash.\",\n      \"type\": \"string\"\n    },\n    \"packageUrl\": {\n      \"description\": \"Enter the packageUrl.\",\n      \"type\": \"string\"\n    },\n    \"stageId\": {\n      \"description\": \"Specify the evaluation report stage.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"stageId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the SearchComponent tool (Status: 200, Content-Type: application/json)
const SearchComponentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains (a) criteria (the search criteria in the request), and (b) results (list of applications with the component specified).\n\nEach result includes applicationId and application name containing the component, the relative and absoluteURLs of the report, component metadata, threat level, and dependency data indicating if the component is a direct/transitive/InnerSource dependency.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **criteria** (Type: object):\n    - **stageId** (Type: string):\n    - **componentIdentifier** (Type: object):\n      - **coordinates** (Type: object):\n        - **Additional Properties**:\n          - **property value** (Type: string):\n      - **format** (Type: string):\n    - **hash** (Type: string):\n    - **packageUrl** (Type: string):\n  - **results** (Type: array):\n    - **Items** (Type: object):\n      - **threatLevel** (Type: integer, int32):\n      - **hash** (Type: string):\n      - **applicationName** (Type: string):\n      - **dependencyData** (Type: object):\n        - **innerSourceData** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: object):\n            - **innerSourceComponentPurl** (Type: string):\n            - **ownerApplicationId** (Type: string):\n            - **ownerApplicationName** (Type: string):\n        - **parentComponentPurls** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: string):\n        - **directDependency** (Type: boolean):\n        - **innerSource** (Type: boolean):\n      - **packageUrl** (Type: string):\n      - **reportHtmlUrl** (Type: string):\n      - **applicationId** (Type: string):\n      - **[cyclic reference]**\n      - **reportUrl** (Type: string):\n"

// NewSearchComponentMCPTool creates the MCP Tool instance for SearchComponent
func NewSearchComponentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SearchComponent",
		"Use this method to retrieve the component details from the application evaluation reports by specifying the component search parameters, format and evaluation stage. You can specify the component search parameters in any one of the 3 ways:<ul><li>SHA1 hash of the component</li><li>Component identifier object containing the coordinates of the component and its format</li><li>packageUrl string</li></ul>Use of wildcards when searching using the GAVEC(coordinates) is supported.\n\nPermissions required: View IQ Elements",
		[]byte(SearchComponentInputSchema),
	)
}

// SearchComponentHandler is the handler function for the SearchComponent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SearchComponentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/search/component", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SearchComponent")
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
