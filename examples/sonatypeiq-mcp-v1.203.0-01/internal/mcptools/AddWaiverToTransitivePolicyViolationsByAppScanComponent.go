package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddWaiverToTransitivePolicyViolationsByAppScanComponent tool
const AddWaiverToTransitivePolicyViolationsByAppScanComponentInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Options for creating policy waivers\",\n      \"properties\": {\n        \"comment\": {\n          \"description\": \"Reason for waiving the violation(s). Must be non-blank.\",\n          \"example\": \"False positive - internal tool approved by security team\",\n          \"type\": \"string\"\n        },\n        \"expireWhenRemediationAvailable\": {\n          \"description\": \"If true, the waiver will automatically expire when a remediation becomes available. Can only be set to true when matcherStrategy is EXACT_COMPONENT.\",\n          \"example\": false,\n          \"type\": \"boolean\"\n        },\n        \"expiryTime\": {\n          \"description\": \"Optional expiration date/time for the waiver in ISO 8601 format. Must be in the future if provided.\",\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"matcherStrategy\": {\n          \"description\": \"Component matching strategy. For Firewall bulk waivers, only EXACT_COMPONENT and ALL_VERSIONS are supported.\",\n          \"enum\": [\n            \"DEFAULT\",\n            \"EXACT_COMPONENT\",\n            \"ALL_COMPONENTS\",\n            \"ALL_VERSIONS\",\n            \"EXACT_COMPONENT\",\n            \"ALL_VERSIONS\"\n          ],\n          \"example\": \"EXACT_COMPONENT\",\n          \"type\": \"string\"\n        },\n        \"waiverReasonId\": {\n          \"description\": \"Optional reference to a pre-defined waiver reason ID\",\n          \"example\": \"waiver-reason-id-123\",\n          \"type\": \"string\"\n        }\n      },\n      \"required\": [\n        \"comment\",\n        \"matcherStrategy\"\n      ],\n      \"type\": \"object\"\n    },\n    \"componentIdentifier\": {\n      \"description\": \"Enter the component identifier of the transitive component on which you want to create a policy waiver.\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"hash\": {\n      \"description\": \"Enter the hash of the transitive component on which you want to create a policy waiver.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Indicates the scope of the waiver that will be created.\",\n      \"enum\": [\n        \"application\"\n      ],\n      \"pattern\": \"application\",\n      \"type\": \"string\"\n    },\n    \"packageUrl\": {\n      \"description\": \"Enter the package URL of the transitive component on which you want to create a policy waiver.\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId (reportId) of the evaluation report that shows the transitive component.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// NewAddWaiverToTransitivePolicyViolationsByAppScanComponentMCPTool creates the MCP Tool instance for AddWaiverToTransitivePolicyViolationsByAppScanComponent
func NewAddWaiverToTransitivePolicyViolationsByAppScanComponentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddWaiverToTransitivePolicyViolationsByAppScanComponent",
		"Use this method to create a policy waiver on a transitive component detected during the specified scan. NOTE: Any one of the input parameters, i.e. component identifier, packageUrl or hash is required. If more than one is provided, the system will pick them in the order specified here.\n\nPermissions required: Waive Policy Violations",
		[]byte(AddWaiverToTransitivePolicyViolationsByAppScanComponentInputSchema),
	)
}

// AddWaiverToTransitivePolicyViolationsByAppScanComponentHandler is the handler function for the AddWaiverToTransitivePolicyViolationsByAppScanComponent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddWaiverToTransitivePolicyViolationsByAppScanComponentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/policyWaivers/transitive/{ownerType}/{ownerId}/{scanId}", args, []string{"ownerId", "ownerType", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddWaiverToTransitivePolicyViolationsByAppScanComponent")
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
