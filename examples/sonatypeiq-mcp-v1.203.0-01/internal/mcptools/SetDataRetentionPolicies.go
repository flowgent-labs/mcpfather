package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetDataRetentionPolicies tool
const SetDataRetentionPoliciesInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON should include the retention policy settings for both application reports and success metrics.\\n\\nPolicy settings for application reports can be specified for each stage of development represented in the example below by additionalProp1. \\nExample values for additionalProp1 are develop, build, stage-release, release, operate \\u0026 continuous monitoring. For application reports created during continuous monitoring use the key continuous-monitoring instead of the stage name. \\u003cul\\u003e\\u003cli\\u003einheritPolicy IS a boolean flag indicating whether the policy is inherited from a parent organization.\\u003c/li\\u003e\\u003cli\\u003eenablePurging IS a boolean flag indicating enabled or disabled status for automatic purging. \\u003c/li\\u003e\\u003cli\\u003emaxCount IS the maximum no. of reports to retain.\\u003c/li\\u003e\\u003cli\\u003emaxAge IS the maximum age that a report is allowed to reach before it is purged. Possible values are days, weeks, months, years.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"applicationReports\": {\n          \"properties\": {\n            \"stages\": {\n              \"additionalProperties\": {\n                \"properties\": {\n                  \"enablePurging\": {\n                    \"type\": \"boolean\"\n                  },\n                  \"inheritPolicy\": {\n                    \"type\": \"boolean\"\n                  },\n                  \"maxAge\": {\n                    \"type\": \"string\"\n                  },\n                  \"maxCount\": {\n                    \"format\": \"int32\",\n                    \"type\": \"integer\"\n                  }\n                },\n                \"type\": \"object\"\n              },\n              \"type\": \"object\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"successMetrics\": {\n          \"properties\": {\n            \"enablePurging\": {\n              \"type\": \"boolean\"\n            },\n            \"inheritPolicy\": {\n              \"type\": \"boolean\"\n            },\n            \"maxAge\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"organizationId\": {\n      \"description\": \"The organizationId for the organization you want to set the data retention policy. Use the organization REST API to retrieve the organizationId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetDataRetentionPoliciesMCPTool creates the MCP Tool instance for SetDataRetentionPolicies
func NewSetDataRetentionPoliciesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetDataRetentionPolicies",
		"Data retention policies help to limit the disk space consumption by removing obsolete data. Use this method to set the retention policies for an organization. Application reports created by continuous monitoring are not affected by the stage retention policy. They appear separately under the key continuous-monitoring.<p>Permissions required: Edit IQ Elements",
		[]byte(SetDataRetentionPoliciesInputSchema),
	)
}

// SetDataRetentionPoliciesHandler is the handler function for the SetDataRetentionPolicies tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetDataRetentionPoliciesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/dataRetentionPolicies/organizations/{organizationId}", args, []string{"organizationId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetDataRetentionPolicies")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
