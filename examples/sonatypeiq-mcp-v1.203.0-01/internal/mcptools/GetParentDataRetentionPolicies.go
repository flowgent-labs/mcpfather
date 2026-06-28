package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetParentDataRetentionPolicies tool
const GetParentDataRetentionPoliciesInputSchema = "{\n  \"properties\": {\n    \"organizationId\": {\n      \"description\": \"The organizationId assigned by IQ Server. Use the organization REST API to retrieve the parent organizationId\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetParentDataRetentionPolicies tool (Status: 200, Content-Type: application/json)
const GetParentDataRetentionPoliciesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response JSON contains the policy settings for both applicationReports and successMetrics. Policy settings for application reports are shown for each stage of development. <ul><li>inheritPolicy IS a boolean flag indicating whether the policy is inherited from a parent organization.</li><li>enablePurging IS a boolean flag indicating if automatic purging is enabled or disabled.</li><li>maxCount IS the maximum no. of reports to retain.</li><li>maxAge IS the maximum age that a report is allowed to reach before it is purged. Possible values are days, weeks, months, years.</li></ul>The latest application report is always retained, regardless of its age.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationReports** (Type: object):\n    - **stages** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: object):\n          - **inheritPolicy** (Type: boolean):\n          - **maxAge** (Type: string):\n          - **maxCount** (Type: integer, int32):\n          - **enablePurging** (Type: boolean):\n  - **successMetrics** (Type: object):\n    - **enablePurging** (Type: boolean):\n    - **inheritPolicy** (Type: boolean):\n    - **maxAge** (Type: string):\n"

// NewGetParentDataRetentionPoliciesMCPTool creates the MCP Tool instance for GetParentDataRetentionPolicies
func NewGetParentDataRetentionPoliciesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetParentDataRetentionPolicies",
		"Data retention policies help to limit the disk space consumption by removing obsolete data. Use this method to inspect the retention policies that are in effect for the parent organization of the given organization. Application reports created by continuous monitoring are not affected by the stage retention policy. They appear separately under the key continuous-monitoring.<p>Permissions required: View IQ Elements",
		[]byte(GetParentDataRetentionPoliciesInputSchema),
	)
}

// GetParentDataRetentionPoliciesHandler is the handler function for the GetParentDataRetentionPolicies tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetParentDataRetentionPoliciesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/dataRetentionPolicies/organizations/{organizationId}/parent", args, []string{"organizationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetParentDataRetentionPolicies")
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
