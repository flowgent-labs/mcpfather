package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetPriorities tool
const GetPrioritiesInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the applicationId.\",\n      \"type\": \"string\"\n    },\n    \"componentNameFilter\": {\n      \"description\": \"Component name to filter by\",\n      \"type\": \"string\"\n    },\n    \"filterOnPolicyActions\": {\n      \"default\": true,\n      \"description\": \"Whether to enable Fail/Warn policy action filter or not\",\n      \"type\": \"boolean\"\n    },\n    \"includeRemediation\": {\n      \"default\": false,\n      \"description\": \"Whether to include remediation type and version for the component or not\",\n      \"type\": \"boolean\"\n    },\n    \"page\": {\n      \"default\": 1,\n      \"description\": \"Current page number.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"default\": 10,\n      \"description\": \"Enter the no. of results that should be visible per page.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPriorities tool (Status: 200, Content-Type: application/json)
const GetPrioritiesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response field " + "\x60" + "priorities" + "\x60" + " returns prioritized components for the specified\napplication ID and scan ID. Each result has relevant component information, reachability\ninformation, policy information, and a priority number, sorted by priority in descending order.\nPagination is supported, and the default page size is 10.\nThe parameter " + "\x60" + "includeRemediation" + "\x60" + " is required for the paginated result to\ninclude remediation information.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **hasAutoWaiversConfigured** (Type: boolean):\n  - **priorities**: Paginated response wrapper (Type: object):\n    - **page**: Current page number (Type: integer, int32):\n    - **pageCount**: Total number of pages (Type: integer, int64):\n    - **pageSize**: Number of items per page (Type: integer, int32):\n    - **results**: List of items for the current page (Type: array):\n      - **Items**: List of items for the current page (Type: object):\n        - **waiverExpirationDetails** (Type: string):\n        - **waivedViolationsCount** (Type: integer, int32):\n        - **highestThreatPolicyName** (Type: string):\n        - **highestReachableThreat** (Type: integer, int32):\n        - **highestThreatPolicyConstraintName** (Type: string):\n        - **displayName** (Type: string):\n        - **hasSameViolationsOnMain** (Type: boolean):\n        - **componentIdentifier** (Type: object):\n          - **coordinates** (Type: object):\n            - **Additional Properties**:\n              - **property value** (Type: string):\n          - **format** (Type: string):\n        - **hasExpiredWaiver** (Type: boolean):\n        - **hasFailActionOnComponent** (Type: boolean):\n        - **hasSoonToExpireWaiver** (Type: boolean):\n        - **remediationType** (Type: string):\n            - Enum: ['next-no-violations', 'next-non-failing', 'next-no-violations-with-dependencies', 'next-non-failing-with-dependencies', 'inner-source-latest-non-breaking', 'inner-source-latest', 'recommended-non-breaking', 'recommended-non-breaking-with-dependencies']\n        - **securityReachable** (Type: boolean):\n        - **hasAutoWaiver** (Type: boolean):\n        - **action** (Type: string):\n        - **highestThreat** (Type: integer, int32):\n        - **dependencyType** (Type: string):\n        - **isAllViolationsWaived** (Type: boolean):\n        - **remediationVersion** (Type: string):\n        - **componentHash** (Type: string):\n        - **priority** (Type: integer, int32):\n    - **total**: Total number of items (Type: integer, int64):\n  - **scanIdFromLatestBuildStageEvaluation** (Type: string):\n"

// NewGetPrioritiesMCPTool creates the MCP Tool instance for GetPriorities
func NewGetPrioritiesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPriorities",
		"Use this method to retrieve all priorities by providing the application ID and scan ID.\n\nPermissions required: View IQ Elements",
		[]byte(GetPrioritiesInputSchema),
	)
}

// GetPrioritiesHandler is the handler function for the GetPriorities tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPrioritiesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/developer/priorities/{applicationId}/{scanId}", args, []string{"applicationId", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPriorities")
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
