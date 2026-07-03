package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetQuarantineList tool
const GetQuarantineListInputSchema = "{\n  \"properties\": {\n    \"asc\": {\n      \"default\": false,\n      \"description\": \"Select the sort order.\",\n      \"type\": \"boolean\"\n    },\n    \"componentName\": {\n      \"description\": \"Enter the component name.\",\n      \"type\": \"string\"\n    },\n    \"page\": {\n      \"default\": 1,\n      \"description\": \"Enter the starting page number for the response.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"default\": 10,\n      \"description\": \"Enter the page size for the response.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"policyId\": {\n      \"description\": \"Enter the list of policy IDs causing the quarantine.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"quarantineTime\": {\n      \"description\": \"Enter the quarantine time of the component.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"repositoryPublicId\": {\n      \"description\": \"Enter the repository public ID of the quarantined component.\",\n      \"type\": \"string\"\n    },\n    \"sortBy\": {\n      \"description\": \"Enter " + "\x60" + "quarantineTime" + "\x60" + " to sort the results by quarantine time.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetQuarantineList tool (Status: 200, Content-Type: application/json)
const GetQuarantineListResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response includes:<ul><li>" + "\x60" + "total" + "\x60" + " is the total number of records this request can return across all pages.</li><li>" + "\x60" + "page" + "\x60" + " is the page number specified in the request.</li><li>" + "\x60" + "pageSize" + "\x60" + " is the page size specified in the request.</li><li>" + "\x60" + "pageCount" + "\x60" + " is the total number of pages this request can return.</li></ul>The " + "\x60" + "results" + "\x60" + " section contains details of each component that has been auto-released. It includes:<ul><li>" + "\x60" + "threatLevel" + "\x60" + " is the threat level of the policy violation.</li><li>" + "\x60" + "policyName" + "\x60" + " is the name of the violated policy.</li><li>" + "\x60" + "quarantined" + "\x60" + " indicates whether the component is quarantined.</li><li>" + "\x60" + "quarantineDate" + "\x60" + " is the date and time when the component was quarantined.</li><li>" + "\x60" + "componentIdentifier" + "\x60" + " is the format and coordinates for the component.</li><li>" + "\x60" + "pathname" + "\x60" + " indicates the component path in the repository.</li><li>" + "\x60" + "displayName" + "\x60" + " is the name and version of the component.</li><li>" + "\x60" + "repositoryId" + "\x60" + " is the ID of the repository where the component is stored.</li><li>" + "\x60" + "repositoryName" + "\x60" + " indicates the repository name where the component is stored.</li><li>" + "\x60" + "hash" + "\x60" + " is the hash of the component.</li><li>" + "\x60" + "matchState" + "\x60" + " indicates the whether the component is an " + "\x60" + "EXACT" + "\x60" + " or " + "\x60" + "SIMILAR" + "\x60" + " match to the known components or is " + "\x60" + "UNKNOWN" + "\x60" + ".</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **pageCount**: Total number of pages (Type: integer, int64):\n  - **pageSize**: Number of items per page (Type: integer, int32):\n  - **results**: List of items for the current page (Type: array):\n    - **Items**: List of items for the current page (Type: object):\n      - **repositoryId** (Type: string):\n      - **threatLevel** (Type: integer, int32):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **repositoryName** (Type: string):\n      - **hash** (Type: string):\n      - **pathname** (Type: string):\n      - **policyName** (Type: string):\n      - **quarantineDate** (Type: string, date-time):\n      - **displayName** (Type: string):\n      - **matchState** (Type: string):\n      - **quarantined** (Type: boolean):\n  - **total**: Total number of items (Type: integer, int64):\n  - **page**: Current page number (Type: integer, int32):\n"

// NewGetQuarantineListMCPTool creates the MCP Tool instance for GetQuarantineList
func NewGetQuarantineListMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetQuarantineList",
		"Use this method to request a list of quarantined components.\n\nPermissions required: View IQ Elements",
		[]byte(GetQuarantineListInputSchema),
	)
}

// GetQuarantineListHandler is the handler function for the GetQuarantineList tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetQuarantineListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/components/quarantined", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetQuarantineList")
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
