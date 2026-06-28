package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetTransitivePolicyWaiversByAppScanComponent tool
const GetTransitivePolicyWaiversByAppScanComponentInputSchema = "{\n  \"properties\": {\n    \"componentIdentifier\": {\n      \"description\": \"Enter the component identifier for the component for which you want to retrieve the waivers on transitive policy violations, for the specified scanId.\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"hash\": {\n      \"description\": \"Enter the hash for the component for which you want to retrieve the waivers on transitive policy violations, for the specified scanId.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. The response will contain the policy violations that are within the scope specified.\",\n      \"enum\": [\n        \"application\"\n      ],\n      \"pattern\": \"application\",\n      \"type\": \"string\"\n    },\n    \"packageUrl\": {\n      \"description\": \"Enter the package URL for the component for which you want to retrieve the waivers on transitive policy violations, for the specified scanId.\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId (reportId) of the scan for which you want to retrieve the waivers on transitive policy violations occurring due the dependencies of a component.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetTransitivePolicyWaiversByAppScanComponent tool (Status: 200, Content-Type: application/json)
const GetTransitivePolicyWaiversByAppScanComponentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a list of waivers on transitive policy violations for the dependencies of the component specified, for the given scanId.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **componentPolicyWaivers** (Type: array):\n    - **Items** (Type: object):\n      - **vulnerabilityId** (Type: string):\n      - **expiryTime** (Type: string, date-time):\n      - **policyId** (Type: string):\n      - **componentUpgradeAvailable** (Type: boolean):\n      - **forContainerImageComponent** (Type: boolean):\n      - **associatedPackageUrl** (Type: string):\n      - **componentName** (Type: string):\n      - **reasonText** (Type: string):\n      - **scopeOwnerType** (Type: string):\n      - **createTime** (Type: string, date-time):\n      - **scopeOwnerName** (Type: string):\n      - **threatLevel** (Type: integer, int32):\n      - **expireWhenRemediationAvailable** (Type: boolean):\n      - **hash** (Type: string):\n      - **forContainerImage** (Type: boolean):\n      - **scopeOwnerId** (Type: string):\n      - **creatorId** (Type: string):\n      - **isObsolete** (Type: boolean):\n      - **policyViolationId** (Type: string):\n      - **matcherStrategy** (Type: string):\n          - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n      - **policyName** (Type: string):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **constraintFacts** (Type: array):\n        - **Items** (Type: object):\n          - **conditionFacts** (Type: array):\n            - **Items** (Type: object):\n              - **conditionIndex** (Type: integer, int32):\n              - **conditionTypeId** (Type: string):\n              - **reason** (Type: string):\n              - **reference** (Type: object):\n                - **value** (Type: string):\n                - **type** (Type: string):\n                    - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n              - **summary** (Type: string):\n              - **triggerJson** (Type: string):\n          - **constraintId** (Type: string):\n          - **constraintName** (Type: string):\n          - **operatorName** (Type: string):\n      - **comment** (Type: string):\n      - **policyWaiverId** (Type: string):\n      - **policyWaiverReasonId** (Type: string):\n      - **constraintFactsJson** (Type: string):\n      - **creatorName** (Type: string):\n      - **displayName** (Type: object):\n        - **name** (Type: string):\n        - **parts** (Type: array):\n          - **Items** (Type: object):\n            - **field** (Type: string):\n            - **value** (Type: string):\n"

// NewGetTransitivePolicyWaiversByAppScanComponentMCPTool creates the MCP Tool instance for GetTransitivePolicyWaiversByAppScanComponent
func NewGetTransitivePolicyWaiversByAppScanComponentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetTransitivePolicyWaiversByAppScanComponent",
		"Use this method to retrieve all waivers on policy violations due to transitive dependencies for a specific component detected in a specific scan. Any one of the input parameters, i.e. componentIdentifier, packageUrl or hash is required. If more than one is provided, the system will pick them in the order specified here.\n\nPermissions required: View IQ Elements",
		[]byte(GetTransitivePolicyWaiversByAppScanComponentInputSchema),
	)
}

// GetTransitivePolicyWaiversByAppScanComponentHandler is the handler function for the GetTransitivePolicyWaiversByAppScanComponent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetTransitivePolicyWaiversByAppScanComponentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyWaivers/transitive/{ownerType}/{ownerId}/{scanId}", args, []string{"ownerId", "ownerType", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetTransitivePolicyWaiversByAppScanComponent")
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
