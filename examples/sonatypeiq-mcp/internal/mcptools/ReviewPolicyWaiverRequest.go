package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the ReviewPolicyWaiverRequest tool
const ReviewPolicyWaiverRequestInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON can include the fields\\u003col\\u003e\\u003cli\\u003estatus. Can be APPROVED or REJECTED\\u003c/li\\u003e\\u003cli\\u003erejectionReason (optional). A text explaining the reason for the rejection., \\u003cli\\u003ecomment (optional, to indicate the reason of the waiver) default value is null\\u003c/li\\u003e\\u003cli\\u003ematcherStrategy (enumeration, required) can have values DEFAULT, EXACT_COMPONENT, ALL_COMPONENTS, ALL_VERSIONS. DEFAULT will match all components if no hash is provided.\\u003c/li\\u003e\\u003cli\\u003eexpiryTime (default null) to set the datetime when the waiver expires.\\u003c/li\\u003e\\u003c/ol\\u003e\\u003cli\\u003eexpireWhenRemediationAvailable (default false) to expire the waiver when a remediation is available.\\u003c/li\\u003e\",\n      \"properties\": {\n        \"comment\": {\n          \"type\": \"string\"\n        },\n        \"expireWhenRemediationAvailable\": {\n          \"type\": \"boolean\"\n        },\n        \"expiryTime\": {\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"matcherStrategy\": {\n          \"enum\": [\n            \"DEFAULT\",\n            \"EXACT_COMPONENT\",\n            \"ALL_COMPONENTS\",\n            \"ALL_VERSIONS\"\n          ],\n          \"type\": \"string\"\n        },\n        \"rejectionReason\": {\n          \"type\": \"string\"\n        },\n        \"status\": {\n          \"type\": \"string\"\n        },\n        \"waiverReasonId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"ownerId\": {\n      \"description\": \"The id for the ownerType provided above. E.g. applicationId if the ownerType is application.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"The scope of the policy waiver request. Possible values are application, organization, repository, repository_manager, repository_container.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    },\n    \"policyWaiverRequestId\": {\n      \"description\": \"The policyWaiverRequestId for the policy waiver request to be approved or rejected.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"ownerId\",\n    \"ownerType\",\n    \"policyWaiverRequestId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ReviewPolicyWaiverRequest tool (Status: 200, Content-Type: application/json)
const ReviewPolicyWaiverRequestResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The updated policy waiver request.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **vulnerabilityId** (Type: string):\n  - **reasonText** (Type: string):\n  - **noteToReviewer** (Type: string):\n  - **rejectionReason** (Type: string):\n  - **requesterId** (Type: string):\n  - **associatedPackageUrl** (Type: string):\n  - **constraintFacts** (Type: array):\n    - **Items** (Type: object):\n      - **operatorName** (Type: string):\n      - **conditionFacts** (Type: array):\n        - **Items** (Type: object):\n          - **summary** (Type: string):\n          - **triggerJson** (Type: string):\n          - **conditionIndex** (Type: integer, int32):\n          - **conditionTypeId** (Type: string):\n          - **reason** (Type: string):\n          - **reference** (Type: object):\n            - **value** (Type: string):\n            - **type** (Type: string):\n                - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n      - **constraintId** (Type: string):\n      - **constraintName** (Type: string):\n  - **scopeOwnerType** (Type: string):\n  - **policyWaiverReasonId** (Type: string):\n  - **status** (Type: string):\n  - **reviewerId** (Type: string):\n  - **expiryTime** (Type: string, date-time):\n  - **hash** (Type: string):\n  - **requesterName** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **expireWhenRemediationAvailable** (Type: boolean):\n  - **policyId** (Type: string):\n  - **policyName** (Type: string):\n  - **policyViolationId** (Type: string):\n  - **constraintFactsJson** (Type: string):\n  - **policyWaiverRequestId** (Type: string):\n  - **isObsolete** (Type: boolean):\n  - **requestTime** (Type: string, date-time):\n  - **displayName** (Type: object):\n    - **name** (Type: string):\n    - **parts** (Type: array):\n      - **Items** (Type: object):\n        - **field** (Type: string):\n        - **value** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n  - **componentName** (Type: string):\n  - **comment** (Type: string):\n  - **matcherStrategy** (Type: string):\n      - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n  - **reviewerName** (Type: string):\n  - **componentUpgradeAvailable** (Type: boolean):\n  - **scopeOwnerName** (Type: string):\n  - **scopeOwnerId** (Type: string):\n"

// NewReviewPolicyWaiverRequestMCPTool creates the MCP Tool instance for ReviewPolicyWaiverRequest
func NewReviewPolicyWaiverRequestMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ReviewPolicyWaiverRequest",
		"Use this method to approve or reject a policy waiver request.\n\nPermissions required: Waive Policy Violations",
		[]byte(ReviewPolicyWaiverRequestInputSchema),
	)
}

// ReviewPolicyWaiverRequestHandler is the handler function for the ReviewPolicyWaiverRequest tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ReviewPolicyWaiverRequestHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/policyWaiverRequests/{ownerType}/{ownerId}/review/{policyWaiverRequestId}", args, []string{"ownerId", "ownerType", "policyWaiverRequestId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ReviewPolicyWaiverRequest")
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
