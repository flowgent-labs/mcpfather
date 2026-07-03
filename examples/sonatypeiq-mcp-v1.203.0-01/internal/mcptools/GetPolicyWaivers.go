package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetPolicyWaivers tool
const GetPolicyWaiversInputSchema = "{\n  \"properties\": {\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. The response will contain waivers that are within the scope specified.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPolicyWaivers tool (Status: 200, Content-Type: application/json)
const GetPolicyWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains waiver details for the specified ownerType and the corresponding ownerId, grouped by the policyWaiverId. The response field 'matcherStrategy' indicates whether the waiver applies to a specific component, or all components that exist at that level of hierarchy (root org, org application), or all versions of the component (past, present, and future). The response fields associatedPackageUrl, displayName, and componentIdentifier are null for waivers on all components and unknown components.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **constraintFactsJson** (Type: string):\n    - **constraintFacts** (Type: array):\n      - **Items** (Type: object):\n        - **constraintName** (Type: string):\n        - **operatorName** (Type: string):\n        - **conditionFacts** (Type: array):\n          - **Items** (Type: object):\n            - **reference** (Type: object):\n              - **type** (Type: string):\n                  - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n              - **value** (Type: string):\n            - **summary** (Type: string):\n            - **triggerJson** (Type: string):\n            - **conditionIndex** (Type: integer, int32):\n            - **conditionTypeId** (Type: string):\n            - **reason** (Type: string):\n        - **constraintId** (Type: string):\n    - **creatorName** (Type: string):\n    - **expiryTime** (Type: string, date-time):\n    - **isObsolete** (Type: boolean):\n    - **vulnerabilityId** (Type: string):\n    - **comment** (Type: string):\n    - **scopeOwnerId** (Type: string):\n    - **componentName** (Type: string):\n    - **creatorId** (Type: string):\n    - **forContainerImage** (Type: boolean):\n    - **createTime** (Type: string, date-time):\n    - **displayName** (Type: object):\n      - **name** (Type: string):\n      - **parts** (Type: array):\n        - **Items** (Type: object):\n          - **field** (Type: string):\n          - **value** (Type: string):\n    - **policyName** (Type: string):\n    - **hash** (Type: string):\n    - **associatedPackageUrl** (Type: string):\n    - **matcherStrategy** (Type: string):\n        - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n    - **policyWaiverId** (Type: string):\n    - **scopeOwnerName** (Type: string):\n    - **forContainerImageComponent** (Type: boolean):\n    - **policyId** (Type: string):\n    - **threatLevel** (Type: integer, int32):\n    - **reasonText** (Type: string):\n    - **componentIdentifier** (Type: object):\n      - **coordinates** (Type: object):\n        - **Additional Properties**:\n          - **property value** (Type: string):\n      - **format** (Type: string):\n    - **componentUpgradeAvailable** (Type: boolean):\n    - **policyViolationId** (Type: string):\n    - **policyWaiverReasonId** (Type: string):\n    - **expireWhenRemediationAvailable** (Type: boolean):\n    - **scopeOwnerType** (Type: string):\n"

// NewGetPolicyWaiversMCPTool creates the MCP Tool instance for GetPolicyWaivers
func NewGetPolicyWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPolicyWaivers",
		"Use this method to retrieve waiver details for all policy waivers for the scope specified. You can specify the scope by using the parameters ownerType and ownerId.\n\nPermissions required: View IQ Elements",
		[]byte(GetPolicyWaiversInputSchema),
	)
}

// GetPolicyWaiversHandler is the handler function for the GetPolicyWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPolicyWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyWaivers/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPolicyWaivers")
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
