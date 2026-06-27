package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetPolicyWaiverRequest tool
const GetPolicyWaiverRequestInputSchema = "{\n  \"properties\": {\n    \"ownerId\": {\n      \"description\": \"The id for the ownerType provided above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"The scope of the policy waiver request. Possible values are application,\\norganization, repository, repository_manager, repository_container.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    },\n    \"policyWaiverRequestId\": {\n      \"description\": \"The policyWaiverRequestId for which you want to retrieve the details.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\",\n    \"policyWaiverRequestId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPolicyWaiverRequest tool (Status: 200, Content-Type: application/json)
const GetPolicyWaiverRequestResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The requested policy waiver request.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **requesterId** (Type: string):\n  - **requesterName** (Type: string):\n  - **reasonText** (Type: string):\n  - **constraintFacts** (Type: array):\n    - **Items** (Type: object):\n      - **constraintName** (Type: string):\n      - **operatorName** (Type: string):\n      - **conditionFacts** (Type: array):\n        - **Items** (Type: object):\n          - **reference** (Type: object):\n            - **value** (Type: string):\n            - **type** (Type: string):\n                - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n          - **summary** (Type: string):\n          - **triggerJson** (Type: string):\n          - **conditionIndex** (Type: integer, int32):\n          - **conditionTypeId** (Type: string):\n          - **reason** (Type: string):\n      - **constraintId** (Type: string):\n  - **isObsolete** (Type: boolean):\n  - **expireWhenRemediationAvailable** (Type: boolean):\n  - **comment** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n  - **reviewerName** (Type: string):\n  - **scopeOwnerId** (Type: string):\n  - **policyViolationId** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **displayName** (Type: object):\n    - **name** (Type: string):\n    - **parts** (Type: array):\n      - **Items** (Type: object):\n        - **field** (Type: string):\n        - **value** (Type: string):\n  - **reviewerId** (Type: string):\n  - **scopeOwnerName** (Type: string):\n  - **hash** (Type: string):\n  - **componentUpgradeAvailable** (Type: boolean):\n  - **policyId** (Type: string):\n  - **matcherStrategy** (Type: string):\n      - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n  - **policyWaiverReasonId** (Type: string):\n  - **policyWaiverRequestId** (Type: string):\n  - **noteToReviewer** (Type: string):\n  - **rejectionReason** (Type: string):\n  - **componentName** (Type: string):\n  - **policyName** (Type: string):\n  - **scopeOwnerType** (Type: string):\n  - **status** (Type: string):\n  - **expiryTime** (Type: string, date-time):\n  - **associatedPackageUrl** (Type: string):\n  - **vulnerabilityId** (Type: string):\n  - **constraintFactsJson** (Type: string):\n  - **requestTime** (Type: string, date-time):\n"

// NewGetPolicyWaiverRequestMCPTool creates the MCP Tool instance for GetPolicyWaiverRequest
func NewGetPolicyWaiverRequestMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPolicyWaiverRequest",
		"Use this method to retrieve policy waiver request details for the policyWaiverRequestId specified.\n\nPermissions required: View IQ Elements",
		[]byte(GetPolicyWaiverRequestInputSchema),
	)
}

// GetPolicyWaiverRequestHandler is the handler function for the GetPolicyWaiverRequest tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPolicyWaiverRequestHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyWaiverRequests/{ownerType}/{ownerId}/{policyWaiverRequestId}", args, []string{"ownerId", "ownerType", "policyWaiverRequestId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPolicyWaiverRequest")
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
