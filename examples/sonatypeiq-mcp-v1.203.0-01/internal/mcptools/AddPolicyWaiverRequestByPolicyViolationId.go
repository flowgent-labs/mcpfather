package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddPolicyWaiverRequestByPolicyViolationId tool
const AddPolicyWaiverRequestByPolicyViolationIdInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON can include the fields\\u003col\\u003e\\u003cli\\u003ecomment (optional, to indicate the reason of the waiver) default value is null\\u003c/li\\u003e\\u003cli\\u003ematcherStrategy (enumeration, required) can have values DEFAULT, EXACT_COMPONENT, ALL_COMPONENTS, ALL_VERSIONS. DEFAULT will match all components if no hash is provided.\\u003c/li\\u003e\\u003cli\\u003eexpiryTime (default null) to set the datetime when the waiver expires.\\u003c/li\\u003e\\u003cli\\u003eexpireWhenRemediationAvailable (default false) to expire the waiver when a remediation is available.\\u003c/li\\u003e\\u003cli\\u003enoteToReviewer (optional) to add a note to the reviewer\\u003c/li\\u003e\\u003c/ol\\u003e\",\n      \"properties\": {\n        \"comment\": {\n          \"type\": \"string\"\n        },\n        \"expireWhenRemediationAvailable\": {\n          \"type\": \"boolean\"\n        },\n        \"expiryTime\": {\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"matcherStrategy\": {\n          \"enum\": [\n            \"DEFAULT\",\n            \"EXACT_COMPONENT\",\n            \"ALL_COMPONENTS\",\n            \"ALL_VERSIONS\"\n          ],\n          \"type\": \"string\"\n        },\n        \"noteToReviewer\": {\n          \"type\": \"string\"\n        },\n        \"waiverReasonId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"ownerId\": {\n      \"description\": \"The id for the ownerType provided above. E.g. applicationId if the ownerType is application.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"The scope of the policy waiver request. Possible values are application, organization, repository, repository_manager, repository_container.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    },\n    \"policyViolationId\": {\n      \"description\": \"The policyViolationId for the policy violation on which you want to create a policy waiver request. Use the Policy Violation REST API or Reports REST API to obtain the policyViolationId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"ownerId\",\n    \"ownerType\",\n    \"policyViolationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddPolicyWaiverRequestByPolicyViolationId tool (Status: 200, Content-Type: application/json)
const AddPolicyWaiverRequestByPolicyViolationIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The new policy waiver request.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **isObsolete** (Type: boolean):\n  - **expireWhenRemediationAvailable** (Type: boolean):\n  - **policyWaiverReasonId** (Type: string):\n  - **reviewerId** (Type: string):\n  - **status** (Type: string):\n  - **displayName** (Type: object):\n    - **name** (Type: string):\n    - **parts** (Type: array):\n      - **Items** (Type: object):\n        - **field** (Type: string):\n        - **value** (Type: string):\n  - **vulnerabilityId** (Type: string):\n  - **constraintFacts** (Type: array):\n    - **Items** (Type: object):\n      - **constraintName** (Type: string):\n      - **operatorName** (Type: string):\n      - **conditionFacts** (Type: array):\n        - **Items** (Type: object):\n          - **conditionIndex** (Type: integer, int32):\n          - **conditionTypeId** (Type: string):\n          - **reason** (Type: string):\n          - **reference** (Type: object):\n            - **type** (Type: string):\n                - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n            - **value** (Type: string):\n          - **summary** (Type: string):\n          - **triggerJson** (Type: string):\n      - **constraintId** (Type: string):\n  - **scopeOwnerType** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n  - **componentUpgradeAvailable** (Type: boolean):\n  - **constraintFactsJson** (Type: string):\n  - **requesterName** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **policyWaiverRequestId** (Type: string):\n  - **requesterId** (Type: string):\n  - **associatedPackageUrl** (Type: string):\n  - **policyName** (Type: string):\n  - **policyId** (Type: string):\n  - **scopeOwnerName** (Type: string):\n  - **expiryTime** (Type: string, date-time):\n  - **componentName** (Type: string):\n  - **policyViolationId** (Type: string):\n  - **reasonText** (Type: string):\n  - **matcherStrategy** (Type: string):\n      - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n  - **rejectionReason** (Type: string):\n  - **hash** (Type: string):\n  - **reviewerName** (Type: string):\n  - **noteToReviewer** (Type: string):\n  - **comment** (Type: string):\n  - **requestTime** (Type: string, date-time):\n  - **scopeOwnerId** (Type: string):\n"

// NewAddPolicyWaiverRequestByPolicyViolationIdMCPTool creates the MCP Tool instance for AddPolicyWaiverRequestByPolicyViolationId
func NewAddPolicyWaiverRequestByPolicyViolationIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddPolicyWaiverRequestByPolicyViolationId",
		"Use this method to create a policy waiver request.\n\nPermissions required: View IQ Elements",
		[]byte(AddPolicyWaiverRequestByPolicyViolationIdInputSchema),
	)
}

// AddPolicyWaiverRequestByPolicyViolationIdHandler is the handler function for the AddPolicyWaiverRequestByPolicyViolationId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddPolicyWaiverRequestByPolicyViolationIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/policyWaiverRequests/{ownerType}/{ownerId}/policyViolation/{policyViolationId}", args, []string{"ownerId", "ownerType", "policyViolationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddPolicyWaiverRequestByPolicyViolationId")
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
