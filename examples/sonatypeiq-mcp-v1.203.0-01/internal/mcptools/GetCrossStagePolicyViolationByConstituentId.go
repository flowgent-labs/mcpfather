package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetCrossStagePolicyViolationByConstituentId tool
const GetCrossStagePolicyViolationByConstituentIdInputSchema = "{\n  \"properties\": {\n    \"constituentId\": {\n      \"description\": \"Enter the violationId. Use the GET method described for the endpoint /api/v2/policyViolations to obtain the policy violationId.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"constituentId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetCrossStagePolicyViolationByConstituentId tool (Status: 200, Content-Type: application/json)
const GetCrossStagePolicyViolationByConstituentIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains violation details for all occurrences of the same policy violation, across multiple stages. stageData indicates the name of the stages where the violation occurred, the scanId/reportId when it was reported and the policy action triggered due to the violation.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **fixTime** (Type: string, date-time):\n  - **openTime** (Type: string, date-time):\n  - **filename** (Type: string):\n  - **legacyViolationTime** (Type: string, date-time):\n  - **displayName** (Type: object):\n    - **name** (Type: string):\n    - **parts** (Type: array):\n      - **Items** (Type: object):\n        - **field** (Type: string):\n        - **value** (Type: string):\n  - **policyName** (Type: string):\n  - **reachabilityStatus** (Type: string):\n      - Enum: ['REACHABLE', 'NON_REACHABLE', 'UNKNOWN']\n  - **applicationName** (Type: string):\n  - **applicationPublicId** (Type: string):\n  - **organizationName** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **hash** (Type: string):\n  - **policyOwner** (Type: object):\n    - **ownerName** (Type: string):\n    - **ownerPublicId** (Type: string):\n    - **ownerType** (Type: string):\n    - **ownerId** (Type: string):\n  - **policyThreatCategory** (Type: string):\n  - **stageData** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: object):\n        - **actionTypeId** (Type: string):\n        - **mostRecentEvaluationTime** (Type: string, date-time):\n        - **mostRecentScanId** (Type: string):\n  - **waiveTime** (Type: string, date-time):\n  - **constraintViolations** (Type: array):\n    - **Items** (Type: object):\n      - **constraintId** (Type: string):\n      - **constraintName** (Type: string):\n      - **reasons** (Type: array):\n        - **Items** (Type: object):\n          - **reason** (Type: string):\n          - **reference** (Type: object):\n            - **type** (Type: string):\n                - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n            - **value** (Type: string):\n  - **policyViolationId** (Type: string):\n  - **policyId** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n"

// NewGetCrossStagePolicyViolationByConstituentIdMCPTool creates the MCP Tool instance for GetCrossStagePolicyViolationByConstituentId
func NewGetCrossStagePolicyViolationByConstituentIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetCrossStagePolicyViolationByConstituentId",
		"A cross-stage policy violation represents an aggregate of all violations of a policy  occurring across multiple stages of an application. Cross-stage policy violations are helpful in performance analysis by determining the time taken to remediate a violation across all stages where it was detected.\nUse this method to retrieve all cross-stage violations, irrespective of the time they were detected.\n\nPermissions required: View IQ Elements",
		[]byte(GetCrossStagePolicyViolationByConstituentIdInputSchema),
	)
}

// GetCrossStagePolicyViolationByConstituentIdHandler is the handler function for the GetCrossStagePolicyViolationByConstituentId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetCrossStagePolicyViolationByConstituentIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyViolations/crossStage", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetCrossStagePolicyViolationByConstituentId")
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
