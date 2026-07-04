package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetMaxAggregationBuckets tool
const GetMaxAggregationBucketsInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetMaxAggregationBuckets tool (Status: 200, Content-Type: application/json)
const GetMaxAggregationBucketsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Maximum aggregation buckets limit\n\n## Response Structure\n\n- Structure (Type: integer):\n    - Example: '10000'\n"

// Response Template for the GetMaxAggregationBuckets tool (Status: 500, Content-Type: application/json)
const GetMaxAggregationBucketsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 500\n\n**Content-Type:** application/json\n\n> Internal server error\n\n## Response Structure\n\n- Structure (Type: object):\n  - **errors** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: string):\n  - **errorMessages** (Type: array):\n    - **Items** (Type: string):\n"

// NewGetMaxAggregationBucketsMCPTool creates the MCP Tool instance for GetMaxAggregationBuckets
func NewGetMaxAggregationBucketsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetMaxAggregationBuckets",
		"Get maximum aggregation buckets - Returns the maximum number of aggregation buckets allowed by the underlying search platform",
		[]byte(GetMaxAggregationBucketsInputSchema),
	)
}

// GetMaxAggregationBucketsHandler is the handler function for the GetMaxAggregationBuckets tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetMaxAggregationBucketsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/searchLimits/maxAggregationBuckets", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetMaxAggregationBuckets")
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
