package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetMetrics tool
const GetMetricsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"applicationIds\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        },\n        \"firstTimePeriod\": {\n          \"type\": \"string\"\n        },\n        \"lastTimePeriod\": {\n          \"type\": \"string\"\n        },\n        \"organizationIds\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        },\n        \"timePeriod\": {\n          \"enum\": [\n            \"WEEK\",\n            \"MONTH\"\n          ],\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetMetrics tool (Status: 200, Content-Type: application/json)
const GetMetricsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Select the media type JSON or csv for the preferred output format.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **applicationPublicId** (Type: string):\n    - **organizationId** (Type: string):\n    - **organizationName** (Type: string):\n    - **aggregations** (Type: array):\n      - **Items** (Type: object):\n        - **evaluationCount** (Type: integer, int32):\n        - **openCountsAtTimePeriodEnd** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: object):\n              - **Additional Properties**:\n                - **property value** (Type: integer, int32):\n        - **timePeriodStart** (Type: string):\n        - **mttrLowThreat** (Type: integer, int64):\n        - **fixedCounts** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: object):\n              - **Additional Properties**:\n                - **property value** (Type: integer, int32):\n        - **mttrCriticalThreat** (Type: integer, int64):\n        - **mttrModerateThreat** (Type: integer, int64):\n        - **discoveredCounts** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: object):\n              - **Additional Properties**:\n                - **property value** (Type: integer, int32):\n        - **mttrSevereThreat** (Type: integer, int64):\n        - **waivedCounts** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: object):\n              - **Additional Properties**:\n                - **property value** (Type: integer, int32):\n    - **applicationId** (Type: string):\n    - **applicationName** (Type: string):\n"

// NewGetMetricsMCPTool creates the MCP Tool instance for GetMetrics
func NewGetMetricsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetMetrics",
		"Use this method to retrieve metrics data such as policy evaluation metrics, violation and remediation metrics aggregated monthly or weekly.\n\nPermissions required: View IQ Elements",
		[]byte(GetMetricsInputSchema),
	)
}

// GetMetricsHandler is the handler function for the GetMetrics tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetMetricsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/reports/metrics", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetMetrics")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
