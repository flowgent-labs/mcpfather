package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPolicyViolations1 tool
const GetPolicyViolations1InputSchema = "{\n  \"properties\": {\n    \"applicationPublicId\": {\n      \"description\": \"Enter the applicationPublicId created at the time of creating the application.\",\n      \"type\": \"string\"\n    },\n    \"includeViolationTimes\": {\n      \"default\": false,\n      \"description\": \"Set to true to include policy violation times (open, legacy, waived, fixed) in the response if set.\",\n      \"type\": \"boolean\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the reportId (scanId) created at the time of evaluating the application.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationPublicId\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPolicyViolations1 tool (Status: 200, Content-Type: application/json)
const GetPolicyViolations1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response fields contain the policy violation data for the reportId (scanId) specified in the method call. The fields corresponding to 'violations' include the violation details for each policy, for the component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **reportTime** (Type: string, date-time):\n  - **reportTitle** (Type: string):\n  - **application** (Type: object):\n    - **publicId** (Type: string):\n    - **contactUserName** (Type: string):\n    - **id** (Type: string):\n    - **name** (Type: string):\n    - **organizationId** (Type: string):\n  - **commitHash** (Type: string):\n  - **components** (Type: array):\n    - **Items** (Type: object):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **displayName** (Type: string):\n      - **hash** (Type: string):\n      - **pathnames** (Type: array):\n        - **Items** (Type: string):\n      - **thirdParty** (Type: boolean):\n      - **dependencyData** (Type: object):\n        - **parentComponentPurls** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: string):\n        - **directDependency** (Type: boolean):\n        - **innerSource** (Type: boolean):\n        - **innerSourceData** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: object):\n            - **innerSourceComponentPurl** (Type: string):\n            - **ownerApplicationId** (Type: string):\n            - **ownerApplicationName** (Type: string):\n      - **matchState** (Type: string):\n      - **proprietary** (Type: boolean):\n      - **sha256** (Type: string):\n      - **originalPurl** (Type: string):\n      - **packageUrl** (Type: string):\n      - **violations** (Type: array):\n        - **Items** (Type: object):\n          - **policyThreatCategory** (Type: string):\n          - **policyThreatLevel** (Type: integer, int32):\n          - **legacyViolationTime** (Type: string, date-time):\n          - **waiveTime** (Type: string, date-time):\n          - **fixTime** (Type: string, date-time):\n          - **policyViolationId** (Type: string):\n          - **waived** (Type: boolean):\n          - **constraints** (Type: array):\n            - **Items** (Type: object):\n              - **conditions** (Type: array):\n                - **Items** (Type: object):\n                  - **conditionReason** (Type: string):\n                  - **conditionSummary** (Type: string):\n              - **constraintId** (Type: string):\n              - **constraintName** (Type: string):\n          - **legacyViolation** (Type: boolean):\n          - **openTime** (Type: string, date-time):\n          - **waivedWithAutoWaiver** (Type: boolean):\n          - **grandfathered** (Type: boolean):\n          - **policyId** (Type: string):\n          - **policyName** (Type: string):\n  - **counts** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: integer, int32):\n  - **initiator** (Type: string):\n"

// NewGetPolicyViolations1MCPTool creates the MCP Tool instance for GetPolicyViolations1
func NewGetPolicyViolations1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPolicyViolations1",
		"Use this method to retrieve the policy violation data generated as a result of an application evaluation, for each component identified in the application evaluation./n/nPermissions required: View IQ Elements",
		[]byte(GetPolicyViolations1InputSchema),
	)
}

// GetPolicyViolations1Handler is the handler function for the GetPolicyViolations1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPolicyViolations1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications/{applicationPublicId}/reports/{scanId}/policy", args, []string{"applicationPublicId", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPolicyViolations1")
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
