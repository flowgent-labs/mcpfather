package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetPolicyViolations tool
const GetPolicyViolationsInputSchema = "{\n  \"properties\": {\n    \"openTimeAfter\": {\n      \"description\": \"Enter the date (format YYYY-MM-DD) from which you want to retrieve the violation details\",\n      \"type\": \"string\"\n    },\n    \"openTimeBefore\": {\n      \"description\": \"Enter the date (format YYYY-MM-DD) until which you want to retrieve the violation details\",\n      \"type\": \"string\"\n    },\n    \"p\": {\n      \"description\": \"Enter the policyIds to obtain the corresponding violation details\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"type\": {\n      \"description\": \"Set one or more policy violation type (active, legacy, waived) to include\",\n      \"items\": {\n        \"default\": \"ACTIVE\",\n        \"enum\": [\n          \"ACTIVE\",\n          \"WAIVED\",\n          \"LEGACY\"\n        ],\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    }\n  },\n  \"required\": [\n    \"p\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPolicyViolations tool (Status: 200, Content-Type: application/json)
const GetPolicyViolationsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of the application that violates the policy/policies and violation details grouped under the policyIds provided. It contains:<ul><li>" + "\x60" + "openTime" + "\x60" + " indicates the date and time when the violation was first detected.</li><li>" + "\x60" + "waiveTime" + "\x60" + " indicates the date and time when the violation was waived.</li><li>" + "\x60" + "legacyTime" + "\x60" + " indicates the date and time when the violation was assigned as a legacy violation.</li><li>" + "\x60" + "reference" + "\x60" + " is the reference data that triggered the violation.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationViolations** (Type: array):\n    - **Items** (Type: object):\n      - **application** (Type: object):\n        - **contactUserName** (Type: string):\n        - **id** (Type: string):\n        - **name** (Type: string):\n        - **organizationId** (Type: string):\n        - **publicId** (Type: string):\n      - **policyViolations** (Type: array):\n        - **Items** (Type: object):\n          - **policyId** (Type: string):\n          - **threatLevel** (Type: integer, int32):\n          - **fixTime** (Type: string, date-time):\n          - **policyName** (Type: string):\n          - **openTime** (Type: string, date-time):\n          - **legacyViolationTime** (Type: string, date-time):\n          - **component** (Type: object):\n            - **componentIdentifier** (Type: object):\n              - **coordinates** (Type: object):\n                - **Additional Properties**:\n                  - **property value** (Type: string):\n              - **format** (Type: string):\n            - **displayName** (Type: string):\n            - **hash** (Type: string):\n            - **originalPurl** (Type: string):\n            - **packageUrl** (Type: string):\n            - **proprietary** (Type: boolean):\n            - **sha256** (Type: string):\n            - **thirdParty** (Type: boolean):\n          - **isWaived** (Type: boolean):\n          - **policyViolationId** (Type: string):\n          - **isLegacy** (Type: boolean):\n          - **constraintViolations** (Type: array):\n            - **Items** (Type: object):\n              - **constraintId** (Type: string):\n              - **constraintName** (Type: string):\n              - **reasons** (Type: array):\n                - **Items** (Type: object):\n                  - **reason** (Type: string):\n                  - **reference** (Type: object):\n                    - **type** (Type: string):\n                        - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n                    - **value** (Type: string):\n          - **reportUrl** (Type: string):\n          - **stageId** (Type: string):\n          - **waiveTime** (Type: string, date-time):\n          - **reportId** (Type: string):\n"

// NewGetPolicyViolationsMCPTool creates the MCP Tool instance for GetPolicyViolations
func NewGetPolicyViolationsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPolicyViolations",
		"Use this method to retrieve policy violation details for a policy/policies. You will need the policyId(s) to retrieve the policy violations details. policyId is available as the response field of the Policies REST API.\n\nPermissions required: View IQ Elements",
		[]byte(GetPolicyViolationsInputSchema),
	)
}

// GetPolicyViolationsHandler is the handler function for the GetPolicyViolations tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPolicyViolationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyViolations", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPolicyViolations")
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
