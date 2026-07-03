package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPolicyViolationDiff tool
const GetPolicyViolationDiffInputSchema = "{\n  \"properties\": {\n    \"applicationPublicId\": {\n      \"description\": \"Enter the applicationPublicId, created at the time of creating the application\",\n      \"type\": \"string\"\n    },\n    \"fromCommit\": {\n      \"description\": \"Enter the commit hash linked to the earlier policy evaluation.\",\n      \"type\": \"string\"\n    },\n    \"fromPolicyEvaluationId\": {\n      \"description\": \"Enter the policy evaluation Id linked to the earlier policy evaluation to compare\",\n      \"type\": \"string\"\n    },\n    \"includeViolationTimes\": {\n      \"default\": false,\n      \"description\": \"Set to true to include policy violation times (open, legacy, waived, fixed) in the response if set.\",\n      \"type\": \"boolean\"\n    },\n    \"toCommit\": {\n      \"description\": \"Enter the commit hash linked to the other (later) policy evaluation to compare.\",\n      \"type\": \"string\"\n    },\n    \"toPolicyEvaluationId\": {\n      \"description\": \"Enter the policy evaluation Id linked to the other (later) policy evaluation to compare\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationPublicId\",\n    \"fromCommit\",\n    \"toCommit\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPolicyViolationDiff tool (Status: 200, Content-Type: application/json)
const GetPolicyViolationDiffResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the violation details grouped under addedViolations, sameViolations and removedViolations for the two policy evaluations being compared.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **sameViolations** (Type: array):\n      - Unique Items: true\n    - **Items** (Type: object):\n      - **threatLevel** (Type: integer, int32):\n      - **constraintViolations** (Type: array):\n        - **Items** (Type: object):\n          - **reasons** (Type: array):\n            - **Items** (Type: object):\n              - **reason** (Type: string):\n              - **reference** (Type: object):\n                - **value** (Type: string):\n                - **type** (Type: string):\n                    - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n          - **constraintId** (Type: string):\n          - **constraintName** (Type: string):\n      - **fixTime** (Type: string, date-time):\n      - **legacyViolationTime** (Type: string, date-time):\n      - **component** (Type: object):\n        - **componentIdentifier** (Type: object):\n          - **coordinates** (Type: object):\n            - **Additional Properties**:\n              - **property value** (Type: string):\n          - **format** (Type: string):\n        - **displayName** (Type: string):\n        - **hash** (Type: string):\n        - **originalPurl** (Type: string):\n        - **packageUrl** (Type: string):\n        - **proprietary** (Type: boolean):\n        - **sha256** (Type: string):\n        - **thirdParty** (Type: boolean):\n      - **policyName** (Type: string):\n      - **waiveTime** (Type: string, date-time):\n      - **openTime** (Type: string, date-time):\n      - **policyId** (Type: string):\n      - **policyViolationId** (Type: string):\n  - **toCommit** (Type: object):\n    - **commitHash** (Type: string):\n    - **reportUrl** (Type: string):\n    - **scanId** (Type: string):\n    - **scanTime** (Type: string, date-time):\n  - **addedViolations** (Type: array):\n      - Unique Items: true\n    - **[cyclic reference]**\n  - **application** (Type: object):\n    - **applicationTags** (Type: array):\n      - **Items** (Type: object):\n        - **applicationId** (Type: string):\n        - **id** (Type: string):\n        - **tagId** (Type: string):\n    - **contactUserName** (Type: string):\n    - **id** (Type: string):\n    - **name** (Type: string):\n    - **organizationId** (Type: string):\n    - **publicId** (Type: string):\n  - **diffTime** (Type: string, date-time):\n  - **[cyclic reference]**\n  - **removedViolations** (Type: array):\n      - Unique Items: true\n    - **[cyclic reference]**\n"

// NewGetPolicyViolationDiffMCPTool creates the MCP Tool instance for GetPolicyViolationDiff
func NewGetPolicyViolationDiffMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPolicyViolationDiff",
		"By configuring Lifecycle with SCM, policy evaluations can be linked to the Git commit hash. Use this method to compare the violations between policy evaluations for 2 commits, by providing the linked commit hashes.\n\nPermissions required: View IQ Elements",
		[]byte(GetPolicyViolationDiffInputSchema),
	)
}

// GetPolicyViolationDiffHandler is the handler function for the GetPolicyViolationDiff tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPolicyViolationDiffHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications/{applicationPublicId}/reports/policyViolations/diff", args, []string{"applicationPublicId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPolicyViolationDiff")
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
