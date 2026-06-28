package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetSimilarWaivers tool
const GetSimilarWaiversInputSchema = "{\n  \"properties\": {\n    \"violationId\": {\n      \"description\": \"Policy violation id to find similar waivers for.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"violationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetSimilarWaivers tool (Status: 200, Content-Type: application/json)
const GetSimilarWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successfully retrieved similar policy waivers for the given policy violation id.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **displayName** (Type: object):\n      - **name** (Type: string):\n      - **parts** (Type: array):\n        - **Items** (Type: object):\n          - **field** (Type: string):\n          - **value** (Type: string):\n    - **policyName** (Type: string):\n    - **hash** (Type: string):\n    - **associatedPackageUrl** (Type: string):\n    - **matcherStrategy** (Type: string):\n        - Enum: ['DEFAULT', 'EXACT_COMPONENT', 'ALL_COMPONENTS', 'ALL_VERSIONS']\n    - **policyWaiverId** (Type: string):\n    - **scopeOwnerName** (Type: string):\n    - **forContainerImageComponent** (Type: boolean):\n    - **policyId** (Type: string):\n    - **threatLevel** (Type: integer, int32):\n    - **reasonText** (Type: string):\n    - **componentIdentifier** (Type: object):\n      - **coordinates** (Type: object):\n        - **Additional Properties**:\n          - **property value** (Type: string):\n      - **format** (Type: string):\n    - **componentUpgradeAvailable** (Type: boolean):\n    - **policyViolationId** (Type: string):\n    - **policyWaiverReasonId** (Type: string):\n    - **expireWhenRemediationAvailable** (Type: boolean):\n    - **scopeOwnerType** (Type: string):\n    - **constraintFactsJson** (Type: string):\n    - **constraintFacts** (Type: array):\n      - **Items** (Type: object):\n        - **constraintName** (Type: string):\n        - **operatorName** (Type: string):\n        - **conditionFacts** (Type: array):\n          - **Items** (Type: object):\n            - **reason** (Type: string):\n            - **reference** (Type: object):\n              - **type** (Type: string):\n                  - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n              - **value** (Type: string):\n            - **summary** (Type: string):\n            - **triggerJson** (Type: string):\n            - **conditionIndex** (Type: integer, int32):\n            - **conditionTypeId** (Type: string):\n        - **constraintId** (Type: string):\n    - **creatorName** (Type: string):\n    - **expiryTime** (Type: string, date-time):\n    - **isObsolete** (Type: boolean):\n    - **vulnerabilityId** (Type: string):\n    - **comment** (Type: string):\n    - **scopeOwnerId** (Type: string):\n    - **componentName** (Type: string):\n    - **creatorId** (Type: string):\n    - **forContainerImage** (Type: boolean):\n    - **createTime** (Type: string, date-time):\n"

// NewGetSimilarWaiversMCPTool creates the MCP Tool instance for GetSimilarWaivers
func NewGetSimilarWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSimilarWaivers",
		"Use this method to retrieve similar policy waivers for the given policy violation id.\n\nPermissions required: View IQ Elements",
		[]byte(GetSimilarWaiversInputSchema),
	)
}

// GetSimilarWaiversHandler is the handler function for the GetSimilarWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSimilarWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyViolations/{violationId}/similarWaivers", args, []string{"violationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSimilarWaivers")
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
