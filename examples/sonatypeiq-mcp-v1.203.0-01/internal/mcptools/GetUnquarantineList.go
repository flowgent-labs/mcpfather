package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetUnquarantineList tool
const GetUnquarantineListInputSchema = "{\n  \"properties\": {\n    \"asc\": {\n      \"default\": true,\n      \"description\": \"Select " + "\x60" + "true" + "\x60" + " to set the sort order to ascending.\",\n      \"type\": \"boolean\"\n    },\n    \"componentName\": {\n      \"description\": \"Enter the component name. When provided, the results will include components with display names (case-insensitive) that match the given name.\",\n      \"type\": \"string\"\n    },\n    \"page\": {\n      \"default\": 1,\n      \"description\": \"Enter the page number.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"default\": 10,\n      \"description\": \"Enter the number of results to be returned for a page.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"policyId\": {\n      \"description\": \"Enter the " + "\x60" + "policyId" + "\x60" + ". When provided, the results will include the components that have a policy violation for the policyId.\",\n      \"type\": \"string\"\n    },\n    \"sortBy\": {\n      \"description\": \"Enter the sort criteria " + "\x60" + "releaseQuarantineTime" + "\x60" + " or " + "\x60" + "quarantineTime" + "\x60" + ".\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetUnquarantineList tool (Status: 200, Content-Type: application/json)
const GetUnquarantineListResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response includes:<ul><li>" + "\x60" + "total" + "\x60" + " is the total number of records this request can return across all pages.</li><li>" + "\x60" + "page" + "\x60" + " is the page number specified in the request.</li><li>" + "\x60" + "pageSize" + "\x60" + " is the page size specified in the request.</li><li>" + "\x60" + "pageCount" + "\x60" + " is the total number of pages this request can return.</li></ul>The " + "\x60" + "results" + "\x60" + " section contains details of each component that has been auto-released. It includes:<ul><li>" + "\x60" + "displayName" + "\x60" + " is the name and version of the component.</li><li>" + "\x60" + "repository" + "\x60" + " indicates the repository name where the component is stored.</li><li>" + "\x60" + "quarantineDate" + "\x60" + " is the date and time when the component was quarantined.</li><li>" + "\x60" + "dateCleared" + "\x60" + " is the date and time when the component was auto-released from quarantine.</li><li>" + "\x60" + "quarantinePolicyViolations" + "\x60" + " will be empty for components that are auto-released.</li><li>" + "\x60" + "componentIdentifier" + "\x60" + " is the format and coordinates for the component.</li><li>" + "\x60" + "pathname" + "\x60" + " indicates the component path in the repository.</li><li>" + "\x60" + "hash" + "\x60" + " is the hash of the component.</li><li>" + "\x60" + "matchState" + "\x60" + " indicates the whether the component is an " + "\x60" + "EXACT" + "\x60" + " or " + "\x60" + "SIMILAR" + "\x60" + " match to the known  components or is " + "\x60" + "UNKNOWN" + "\x60" + ".</li><li>" + "\x60" + "repositoryId" + "\x60" + " is the ID of the repository where the component is stored.</li><li>" + "\x60" + "quarantined" + "\x60" + " indicates whether the component is quarantined.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **results**: List of items for the current page (Type: array):\n    - **Items**: List of items for the current page (Type: object):\n      - **displayName** (Type: string):\n      - **matchState** (Type: string):\n      - **quarantinePolicyViolations** (Type: array):\n        - **Items** (Type: object):\n          - **fixTime** (Type: string, date-time):\n          - **policyId** (Type: string):\n          - **openTime** (Type: string, date-time):\n          - **policyName** (Type: string):\n          - **waiveTime** (Type: string, date-time):\n          - **policyViolationId** (Type: string):\n          - **threatLevel** (Type: integer, int32):\n          - **legacyViolationTime** (Type: string, date-time):\n          - **constraintViolations** (Type: array):\n            - **Items** (Type: object):\n              - **constraintId** (Type: string):\n              - **constraintName** (Type: string):\n              - **reasons** (Type: array):\n                - **Items** (Type: object):\n                  - **reason** (Type: string):\n                  - **reference** (Type: object):\n                    - **type** (Type: string):\n                        - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n                    - **value** (Type: string):\n      - **repository** (Type: string):\n      - **hash** (Type: string):\n      - **pathname** (Type: string):\n      - **quarantineDate** (Type: string, date-time):\n      - **quarantined** (Type: boolean):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **repositoryId** (Type: string):\n      - **dateCleared** (Type: string, date-time):\n  - **total**: Total number of items (Type: integer, int64):\n  - **page**: Current page number (Type: integer, int32):\n  - **pageCount**: Total number of pages (Type: integer, int64):\n  - **pageSize**: Number of items per page (Type: integer, int32):\n"

// NewGetUnquarantineListMCPTool creates the MCP Tool instance for GetUnquarantineList
func NewGetUnquarantineListMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUnquarantineList",
		"Use this method to retrieve the details of components that are auto-released from quarantine.\n\nPermissions required: View IQ Elements",
		[]byte(GetUnquarantineListInputSchema),
	)
}

// GetUnquarantineListHandler is the handler function for the GetUnquarantineList tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUnquarantineListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/components/autoReleasedFromQuarantine", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetUnquarantineList")
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
