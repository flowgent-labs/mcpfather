package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetApplicableWaivers tool
const GetApplicableWaiversInputSchema = "{\n  \"properties\": {\n    \"violationId\": {\n      \"description\": \"Enter the policy violationId for which you want to obtain the applicable waivers.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"violationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetApplicableWaivers tool (Status: 200, Content-Type: application/json)
const GetApplicableWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains details for all applicable waivers for the " + "\x60" + "violationId" + "\x60" + " specified. It is grouped under 'activeWaivers' and 'expiredWaivers'. " + "\x60" + "scope" + "\x60" + " indicates the scope of the applicable waiver. Possible values for the enum field " + "\x60" + "matcherStrategy" + "\x60" + " are EXACT_COMPONENT, ALL_COMPONENTS, ALL_VERSIONS).\n\n" + "\x60" + "reference" + "\x60" + " shows the reference data that triggered the violation. " + "\x60" + "componentUpgradeAvailable" + "\x60" + " indicates if a non-violating version of the component is available to remediate the violation.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **activeWaivers** (Type: array):\n    - **Items** (Type: object):\n      - **associatedPackageUrl** (Type: string):\n      - **matcherStrategy** (Type: string):\n          - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n      - **policyWaiverId** (Type: string):\n      - **scopeOwnerName** (Type: string):\n      - **forContainerImageComponent** (Type: boolean):\n      - **policyId** (Type: string):\n      - **threatLevel** (Type: integer, int32):\n      - **reasonText** (Type: string):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **componentUpgradeAvailable** (Type: boolean):\n      - **policyViolationId** (Type: string):\n      - **policyWaiverReasonId** (Type: string):\n      - **expireWhenRemediationAvailable** (Type: boolean):\n      - **scopeOwnerType** (Type: string):\n      - **constraintFactsJson** (Type: string):\n      - **constraintFacts** (Type: array):\n        - **Items** (Type: object):\n          - **operatorName** (Type: string):\n          - **conditionFacts** (Type: array):\n            - **Items** (Type: object):\n              - **summary** (Type: string):\n              - **triggerJson** (Type: string):\n              - **conditionIndex** (Type: integer, int32):\n              - **conditionTypeId** (Type: string):\n              - **reason** (Type: string):\n              - **reference** (Type: object):\n                - **type** (Type: string):\n                    - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n                - **value** (Type: string):\n          - **constraintId** (Type: string):\n          - **constraintName** (Type: string):\n      - **creatorName** (Type: string):\n      - **expiryTime** (Type: string, date-time):\n      - **isObsolete** (Type: boolean):\n      - **vulnerabilityId** (Type: string):\n      - **comment** (Type: string):\n      - **scopeOwnerId** (Type: string):\n      - **componentName** (Type: string):\n      - **creatorId** (Type: string):\n      - **forContainerImage** (Type: boolean):\n      - **createTime** (Type: string, date-time):\n      - **displayName** (Type: object):\n        - **name** (Type: string):\n        - **parts** (Type: array):\n          - **Items** (Type: object):\n            - **field** (Type: string):\n            - **value** (Type: string):\n      - **policyName** (Type: string):\n      - **hash** (Type: string):\n  - **expiredWaivers** (Type: array):\n    - **[cyclic reference]**\n"

// NewGetApplicableWaiversMCPTool creates the MCP Tool instance for GetApplicableWaivers
func NewGetApplicableWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplicableWaivers",
		"Use this method to obtain all existing waivers that are applicable to a policy violation. A waiver is considered as 'applicable' if it matches the following conditions:<ul><li>The policyId for the policy violation matches the policyId associated with the waiver</li><li>The violated policy conditions match the policy conditions of the waiver/li><li>The waiver scope matches the violating component</li></ul>\n\nPermissions required: View IQ Elements",
		[]byte(GetApplicableWaiversInputSchema),
	)
}

// GetApplicableWaiversHandler is the handler function for the GetApplicableWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicableWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyViolations/{violationId}/applicableWaivers", args, []string{"violationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplicableWaivers")
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
