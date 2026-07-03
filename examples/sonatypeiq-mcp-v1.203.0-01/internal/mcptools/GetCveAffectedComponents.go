package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetCveAffectedComponents tool
const GetCveAffectedComponentsInputSchema = "{\n  \"properties\": {\n    \"cveId\": {\n      \"description\": \"CVE identifier(s). Can be specified multiple times for multiple CVEs.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"pageNumber\": {\n      \"default\": 1,\n      \"description\": \"Page number (1-indexed, minimum: 1, default: 1)\",\n      \"format\": \"int32\",\n      \"minimum\": 1,\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"default\": 10,\n      \"description\": \"Number of items per page (1-1000, default: 10)\",\n      \"format\": \"int32\",\n      \"maximum\": 1000,\n      \"minimum\": 1,\n      \"type\": \"integer\"\n    },\n    \"sortBy\": {\n      \"description\": \"Sort field: applicationName, applicationId, componentName, evaluationDate, stage, activeWaiver, violating, cveId. When not specified, sorts by applicationName (asc), then componentName (asc), then cveId (asc)\",\n      \"enum\": [\n        \"APPLICATION_NAME\",\n        \"COMPONENT_NAME\",\n        \"EVALUATION_DATE\",\n        \"STAGE\",\n        \"APPLICATION_ID\",\n        \"ACTIVE_WAIVER\",\n        \"VIOLATING\",\n        \"CVE_ID\"\n      ],\n      \"type\": \"string\"\n    },\n    \"sortOrder\": {\n      \"default\": \"asc\",\n      \"description\": \"Sort order: asc or desc, default: asc\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"cveId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetCveAffectedComponentsMCPTool creates the MCP Tool instance for GetCveAffectedComponents
func NewGetCveAffectedComponentsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCveAffectedComponents",
		"Retrieve paginated list of applications containing components affected by one or more CVEs. Multiple CVE IDs can be specified using multiple cveId query parameters (e.g., ?cveId=CVE-2025-1&cveId=CVE-2025-2). Default page number is 1, default page size is 10. Results can be sorted by any column. Default sorting (when sortBy is not specified): applicationName (asc), then componentName (asc), then cveId (asc). When sortBy is explicitly specified, only single-field sorting is applied with the specified sortOrder (default: asc). <p>Permissions Required: View IQ Elements",
		[]byte(GetCveAffectedComponentsInputSchema),
	)
}

// GetCveAffectedComponentsHandler is the handler function for the GetCveAffectedComponents tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCveAffectedComponentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/componentSearch/cveAffectedComponents", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetCveAffectedComponents")
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
