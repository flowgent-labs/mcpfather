package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetComponentEvaluation tool
const GetComponentEvaluationInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal applicationId (same as that sent in the POST request (step 1))\",\n      \"type\": \"string\"\n    },\n    \"resultId\": {\n      \"description\": \"Enter the resultId obtained from the POST response (step 1) used for component evaluation.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"resultId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetComponentEvaluation tool (Status: 200, Content-Type: application/json)
const GetComponentEvaluationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains details for the policy evaluation request including submitted date, evaluation date, applicationId and the results of the evaluation for the component(s).\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationId** (Type: string):\n  - **errorMessage** (Type: string):\n  - **evaluationDate** (Type: string, date-time):\n  - **isError** (Type: boolean):\n  - **results** (Type: array):\n    - **Items** (Type: object):\n      - **policyData** (Type: object):\n        - **policyViolations** (Type: array):\n          - **Items** (Type: object):\n            - **fixTime** (Type: string, date-time):\n            - **legacyViolationTime** (Type: string, date-time):\n            - **openTime** (Type: string, date-time):\n            - **constraintViolations** (Type: array):\n              - **Items** (Type: object):\n                - **constraintId** (Type: string):\n                - **constraintName** (Type: string):\n                - **reasons** (Type: array):\n                  - **Items** (Type: object):\n                    - **reason** (Type: string):\n                    - **reference** (Type: object):\n                      - **value** (Type: string):\n                      - **type** (Type: string):\n                          - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n            - **policyId** (Type: string):\n            - **policyName** (Type: string):\n            - **policyViolationId** (Type: string):\n            - **threatLevel** (Type: integer, int32):\n            - **waiveTime** (Type: string, date-time):\n      - **projectData** (Type: object):\n        - **lastReleaseDate** (Type: string, date-time):\n        - **projectMetadata** (Type: object):\n          - **description** (Type: string):\n          - **organization** (Type: string):\n        - **sourceControlManagement** (Type: object):\n          - **scmUrl** (Type: string):\n          - **scmDetails** (Type: object):\n            - **commitsPerMonth** (Type: integer, int32):\n            - **uniqueDevsPerMonth** (Type: integer, int32):\n          - **scmMetadata** (Type: object):\n            - **stars** (Type: integer, int32):\n            - **forks** (Type: integer, int32):\n        - **firstReleaseDate** (Type: string, date-time):\n      - **licenseData** (Type: object):\n        - **observedLicenses** (Type: array):\n          - **Items** (Type: object):\n            - **licenseId** (Type: string):\n            - **licenseName** (Type: string):\n        - **overriddenLicenses** (Type: array):\n          - **[cyclic reference]**\n        - **status** (Type: string):\n        - **declaredLicenses** (Type: array):\n          - **[cyclic reference]**\n        - **effectiveLicenses** (Type: array):\n          - **[cyclic reference]**\n      - **relativePopularity** (Type: integer, int32, nullable):\n          - Nullable: true\n      - **hygieneRating** (Type: string, nullable):\n          - Nullable: true\n      - **integrityRating** (Type: string, nullable):\n          - Nullable: true\n      - **securityData** (Type: object):\n        - **securityIssues** (Type: array):\n          - **Items** (Type: object):\n            - **analysis** (Type: object):\n              - **response** (Type: string):\n              - **state** (Type: string):\n              - **detail** (Type: string):\n              - **justification** (Type: string):\n            - **source** (Type: string):\n            - **threatCategory** (Type: string):\n            - **url** (Type: string):\n            - **status** (Type: string):\n            - **cwe** (Type: string):\n            - **severity** (Type: number, float):\n            - **cvssVector** (Type: string):\n            - **cvssVectorSource** (Type: string):\n            - **reference** (Type: string):\n      - **catalogDate** (Type: string, date-time):\n      - **component** (Type: object):\n        - **displayName** (Type: string):\n        - **hash** (Type: string):\n        - **originalPurl** (Type: string):\n        - **packageUrl** (Type: string):\n        - **proprietary** (Type: boolean):\n        - **sha256** (Type: string):\n        - **thirdParty** (Type: boolean):\n        - **componentIdentifier** (Type: object):\n          - **coordinates** (Type: object):\n            - **Additional Properties**:\n              - **property value** (Type: string):\n          - **format** (Type: string):\n      - **matchState** (Type: string):\n  - **submittedDate** (Type: string, date-time):\n"

// NewGetComponentEvaluationMCPTool creates the MCP Tool instance for GetComponentEvaluation
func NewGetComponentEvaluationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetComponentEvaluation",
		"This is step 2 of the policy evaluation process for components. Use the resultId obtained from the POST response for the corresponding applicationId. \n\nPermissions Required: Evaluate Components ",
		[]byte(GetComponentEvaluationInputSchema),
	)
}

// GetComponentEvaluationHandler is the handler function for the GetComponentEvaluation tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetComponentEvaluationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/evaluation/applications/{applicationId}/results/{resultId}", args, []string{"applicationId", "resultId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetComponentEvaluation")
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
