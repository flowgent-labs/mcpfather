package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetComponentsWithWaivers tool
const GetComponentsWithWaiversInputSchema = "{\n  \"properties\": {\n    \"format\": {\n      \"description\": \"Enter the format/ecosystem of the component\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetComponentsWithWaivers tool (Status: 200, Content-Type: application/json)
const GetComponentsWithWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The JSON response contains waivers grouped by application components and repository components. Waived violations for application components are listed per stage. Waived violations for repository components are listed in the Proxy stage. <p>The component hash is null if the waiver applies to all components or all versions of a component. It is truncated and meant to be used as an identifier to other REST API calls and not for use as checksum. <p>" + "\x60" + "isObsolete" + "\x60" + " indicates if a waived violation does not have a valid waiver information. This could happen when a waiver has been removed and the report has not been re-evaluated.<p>" + "\x60" + "matcherStrategy" + "\x60" + " can have values EXACT_COMPONENT, ALL_COMPONENTS, ALL_VERSIONS. <p>The response fields " + "\x60" + "associatedPackageUrl" + "\x60" + ", " + "\x60" + "componentIdentifier" + "\x60" + " and " + "\x60" + "displayName" + "\x60" + " are returned only if the waiver is of type ALL_VERSIONS OR EXACT_COMPONENTS and the component is not an unknown component .\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationWaivers** (Type: array):\n    - **Items** (Type: object):\n      - **application** (Type: object):\n        - **name** (Type: string):\n        - **organizationId** (Type: string):\n        - **publicId** (Type: string):\n        - **contactUserName** (Type: string):\n        - **id** (Type: string):\n      - **stages** (Type: array):\n        - **Items** (Type: object):\n          - **componentPolicyViolations** (Type: array):\n            - **Items** (Type: object):\n              - **waivedPolicyViolations** (Type: array):\n                - **Items** (Type: object):\n                  - **openTime** (Type: string, date-time):\n                  - **policyId** (Type: string):\n                  - **policyName** (Type: string):\n                  - **policyWaiver** (Type: object):\n                    - **forContainerImageComponent** (Type: boolean):\n                    - **componentIdentifier** (Type: object):\n                      - **coordinates** (Type: object):\n                        - **Additional Properties**:\n                          - **property value** (Type: string):\n                      - **format** (Type: string):\n                    - **expiryTime** (Type: string, date-time):\n                    - **reasonText** (Type: string):\n                    - **comment** (Type: string):\n                    - **vulnerabilityId** (Type: string):\n                    - **forContainerImage** (Type: boolean):\n                    - **componentUpgradeAvailable** (Type: boolean):\n                    - **policyWaiverId** (Type: string):\n                    - **displayName** (Type: object):\n                      - **name** (Type: string):\n                      - **parts** (Type: array):\n                        - **Items** (Type: object):\n                          - **value** (Type: string):\n                          - **field** (Type: string):\n                    - **creatorId** (Type: string):\n                    - **creatorName** (Type: string):\n                    - **policyWaiverReasonId** (Type: string):\n                    - **matcherStrategy** (Type: string):\n                        - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n                    - **scopeOwnerName** (Type: string):\n                    - **constraintFacts** (Type: array):\n                      - **Items** (Type: object):\n                        - **operatorName** (Type: string):\n                        - **conditionFacts** (Type: array):\n                          - **Items** (Type: object):\n                            - **conditionIndex** (Type: integer, int32):\n                            - **conditionTypeId** (Type: string):\n                            - **reason** (Type: string):\n                            - **reference** (Type: object):\n                              - **value** (Type: string):\n                              - **type** (Type: string):\n                                  - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n                            - **summary** (Type: string):\n                            - **triggerJson** (Type: string):\n                        - **constraintId** (Type: string):\n                        - **constraintName** (Type: string):\n                    - **hash** (Type: string):\n                    - **threatLevel** (Type: integer, int32):\n                    - **policyId** (Type: string):\n                    - **scopeOwnerType** (Type: string):\n                    - **componentName** (Type: string):\n                    - **associatedPackageUrl** (Type: string):\n                    - **policyName** (Type: string):\n                    - **scopeOwnerId** (Type: string):\n                    - **createTime** (Type: string, date-time):\n                    - **isObsolete** (Type: boolean):\n                    - **expireWhenRemediationAvailable** (Type: boolean):\n                    - **constraintFactsJson** (Type: string):\n                    - **policyViolationId** (Type: string):\n                  - **threatLevel** (Type: integer, int32):\n                  - **policyViolationId** (Type: string):\n                  - **waiveTime** (Type: string, date-time):\n                  - **fixTime** (Type: string, date-time):\n                  - **legacyViolationTime** (Type: string, date-time):\n                  - **constraintViolations** (Type: array):\n                    - **Items** (Type: object):\n                      - **reasons** (Type: array):\n                        - **Items** (Type: object):\n                          - **reason** (Type: string):\n                          - **[cyclic reference]**\n                      - **constraintId** (Type: string):\n                      - **constraintName** (Type: string):\n              - **component** (Type: object):\n                - **originalPurl** (Type: string):\n                - **packageUrl** (Type: string):\n                - **proprietary** (Type: boolean):\n                - **sha256** (Type: string):\n                - **thirdParty** (Type: boolean):\n                - **[cyclic reference]**\n                - **displayName** (Type: string):\n                - **hash** (Type: string):\n          - **stageId** (Type: string):\n  - **repositoryWaivers** (Type: array):\n    - **Items** (Type: object):\n      - **stages** (Type: array):\n        - **[cyclic reference]**\n      - **repository** (Type: object):\n        - **publicId** (Type: string):\n        - **quarantineEnabled** (Type: boolean):\n        - **repositoryId** (Type: string):\n        - **type** (Type: string):\n        - **auditEnabled** (Type: boolean):\n        - **format** (Type: string):\n        - **namespaceConfusionProtectionEnabled** (Type: boolean):\n        - **policyCompliantComponentSelectionEnabled** (Type: boolean):\n"

// NewGetComponentsWithWaiversMCPTool creates the MCP Tool instance for GetComponentsWithWaivers
func NewGetComponentsWithWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetComponentsWithWaivers",
		"Use this method to retrieve existing policy waivers by components. For an up-to-date response, ensure that all application and repository reports are current and contain the most recent re-evaluation data.<p>You can specify the format/ecosystem of the component for a filtered result. <p>Permissions required: View IQ Elements and access to the specific applications and repositories ",
		[]byte(GetComponentsWithWaiversInputSchema),
	)
}

// GetComponentsWithWaiversHandler is the handler function for the GetComponentsWithWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetComponentsWithWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/reports/components/waivers", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetComponentsWithWaivers")
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
