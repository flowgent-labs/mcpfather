package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetPolicyWaiver tool
const GetPolicyWaiverInputSchema = "{\n  \"properties\": {\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. The response will contain the details for waivers within the scope.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    },\n    \"policyWaiverId\": {\n      \"description\": \"Enter the policyWaiverId for which you want to retrieve the waiver details.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\",\n    \"policyWaiverId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPolicyWaiver tool (Status: 200, Content-Type: application/json)
const GetPolicyWaiverResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains waiver details corresponding to the policy waiverId specified.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **associatedPackageUrl** (Type: string):\n  - **matcherStrategy** (Type: string):\n      - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n  - **policyWaiverId** (Type: string):\n  - **scopeOwnerName** (Type: string):\n  - **forContainerImageComponent** (Type: boolean):\n  - **policyId** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n  - **reasonText** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **componentUpgradeAvailable** (Type: boolean):\n  - **policyViolationId** (Type: string):\n  - **policyWaiverReasonId** (Type: string):\n  - **expireWhenRemediationAvailable** (Type: boolean):\n  - **scopeOwnerType** (Type: string):\n  - **constraintFactsJson** (Type: string):\n  - **constraintFacts** (Type: array):\n    - **Items** (Type: object):\n      - **constraintName** (Type: string):\n      - **operatorName** (Type: string):\n      - **conditionFacts** (Type: array):\n        - **Items** (Type: object):\n          - **summary** (Type: string):\n          - **triggerJson** (Type: string):\n          - **conditionIndex** (Type: integer, int32):\n          - **conditionTypeId** (Type: string):\n          - **reason** (Type: string):\n          - **reference** (Type: object):\n            - **type** (Type: string):\n                - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n            - **value** (Type: string):\n      - **constraintId** (Type: string):\n  - **creatorName** (Type: string):\n  - **expiryTime** (Type: string, date-time):\n  - **isObsolete** (Type: boolean):\n  - **vulnerabilityId** (Type: string):\n  - **comment** (Type: string):\n  - **scopeOwnerId** (Type: string):\n  - **componentName** (Type: string):\n  - **creatorId** (Type: string):\n  - **forContainerImage** (Type: boolean):\n  - **createTime** (Type: string, date-time):\n  - **displayName** (Type: object):\n    - **parts** (Type: array):\n      - **Items** (Type: object):\n        - **value** (Type: string):\n        - **field** (Type: string):\n    - **name** (Type: string):\n  - **policyName** (Type: string):\n  - **hash** (Type: string):\n"

// NewGetPolicyWaiverMCPTool creates the MCP Tool instance for GetPolicyWaiver
func NewGetPolicyWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPolicyWaiver",
		"Use this method to retrieve waiver details for the waiverId specified.\n\nPermissions required: View IQ Elements",
		[]byte(GetPolicyWaiverInputSchema),
	)
}

// GetPolicyWaiverHandler is the handler function for the GetPolicyWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPolicyWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyWaivers/{ownerType}/{ownerId}/{policyWaiverId}", args, []string{"ownerId", "ownerType", "policyWaiverId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPolicyWaiver")
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
