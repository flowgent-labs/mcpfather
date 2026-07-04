package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetReportHistoryForApplication tool
const GetReportHistoryForApplicationInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal application Id. You can use the Applications REST API to get the internal application Id. \",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"description\": \"Enter the exact no. of most recent reports to retrieve.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"stage\": {\n      \"description\": \"Enter the specific stage, for which you want retrieve the scan history, e.g. 'build' \",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetReportHistoryForApplication tool (Status: 200, Content-Type: application/json)
const GetReportHistoryForApplicationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains evaluation details, embeddable link and URLs to view the reports in pdf and html formats. \n\n<ul><li><code>stage</code> indicates the stage at which policy evaluation was performed, such as 'develop', 'build' and 'release'.</li><li><code>latestReportHtmlUrl</code> is a relative link to view the most recent evaluation report.</li><li><code>reportPdfUrl</code> and <code>reportHtmlUrl</code> are links to view the pdf version of the report.</li><li><code>reportDataUrl</code> is a link to view the most recent report data.</li><li><code>scanId</code> is the Id associated with the evaluation report.</li><li><code>isReevaluation</code> indicates whether this policy evaluation is a re-evaluation of an older policy evaluation.</li><li><code>isForMonitoring</code> indicates whether this policy evaluation was triggered by continuous monitoring.</li><li><code>commitHash</code> is the hash string that identifies a specific commit in the source control system.</li><li><code>scanTriggerType</code> indicates the type of scan used for this evaluation, such as WEB_UI.</li><li><code>affectedComponentCount</code> is the number of components identified in this evaluation.</li><li><code>criticalComponentCount</code>, <code>severeComponentCount</code>, <code>moderateComponentCount</code> indicate the number of components with critical, severe and moderate policy violations respectively.</li><li><code>criticalPolicyViolationCount</code>, <code>severePolicyViolationCount</code>, <code>moderatePolicyViolationCount</code> indicate the number of critical, severe and moderate policy violations respectively.</li><li><code>policyEvaluationResult</code> contains details on the policy violation such as, coordinates of the violating component and the specific policy constraints that are violated.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationId** (Type: string):\n  - **reports** (Type: array):\n    - **Items** (Type: object):\n      - **scanId** (Type: string):\n      - **commitHash** (Type: string):\n      - **evaluationDate** (Type: string, date-time):\n      - **latestReportHtmlUrl** (Type: string):\n      - **reportHtmlUrl** (Type: string):\n      - **policyEvaluationResult** (Type: object):\n        - **moderateComponentCount** (Type: integer, int32):\n        - **alerts** (Type: array):\n          - **Items** (Type: object):\n            - **actions** (Type: array):\n              - **Items** (Type: object):\n                - **targetType** (Type: string):\n                - **actionTypeId** (Type: string):\n                - **target** (Type: string):\n            - **trigger** (Type: object):\n              - **componentFacts** (Type: array):\n                - **Items** (Type: object):\n                  - **displayName** (Type: object):\n                    - **name** (Type: string):\n                    - **parts** (Type: array):\n                      - **Items** (Type: object):\n                        - **field** (Type: string):\n                        - **value** (Type: string):\n                  - **hash** (Type: string):\n                  - **pathnames** (Type: array):\n                    - **Items** (Type: string):\n                  - **componentIdentifier** (Type: object):\n                    - **coordinates** (Type: object):\n                      - **Additional Properties**:\n                        - **property value** (Type: string):\n                    - **format** (Type: string):\n                  - **constraintFacts** (Type: array):\n                    - **Items** (Type: object):\n                      - **conditionFacts** (Type: array):\n                        - **Items** (Type: object):\n                          - **conditionTypeId** (Type: string):\n                          - **reason** (Type: string):\n                          - **reference** (Type: object):\n                            - **value** (Type: string):\n                            - **type** (Type: string):\n                                - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n                          - **summary** (Type: string):\n                          - **triggerJson** (Type: string):\n                          - **conditionIndex** (Type: integer, int32):\n                      - **constraintId** (Type: string):\n                      - **constraintName** (Type: string):\n                      - **operatorName** (Type: string):\n              - **policyId** (Type: string):\n              - **policyName** (Type: string):\n              - **policyViolationId** (Type: string):\n              - **threatLevel** (Type: integer, int32):\n        - **criticalComponentCount** (Type: integer, int32):\n        - **affectedComponentCount** (Type: integer, int32):\n        - **moderatePolicyViolationCount** (Type: integer, int32):\n        - **criticalSastPolicyViolationCount** (Type: integer, int32):\n        - **legacyViolationCount** (Type: integer, int32):\n        - **totalComponentCount** (Type: integer, int32):\n        - **sastAlerts** (Type: array):\n          - **[cyclic reference]**\n        - **severePolicyViolationCount** (Type: integer, int32):\n        - **grandfatheredPolicyViolationCount** (Type: integer, int32):\n        - **severeComponentCount** (Type: integer, int32):\n        - **severeSastPolicyViolationCount** (Type: integer, int32):\n        - **criticalPolicyViolationCount** (Type: integer, int32):\n        - **moderateSastPolicyViolationCount** (Type: integer, int32):\n        - **totalSastFindingCount** (Type: integer, int32):\n      - **scanTriggerType** (Type: string):\n      - **isForMonitoring** (Type: boolean):\n      - **isReevaluation** (Type: boolean):\n      - **scanTriggerTypeDisplayName** (Type: string):\n      - **applicationId** (Type: string):\n      - **reportPdfUrl** (Type: string):\n      - **scannerVersion** (Type: string):\n      - **reportDataUrl** (Type: string):\n      - **stage** (Type: string):\n      - **policyEvaluationId** (Type: string):\n      - **scanTriggerInternal** (Type: boolean):\n      - **embeddableReportHtmlUrl** (Type: string):\n"

// NewGetReportHistoryForApplicationMCPTool creates the MCP Tool instance for GetReportHistoryForApplication
func NewGetReportHistoryForApplicationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetReportHistoryForApplication",
		"Use this method to retrieve previous application scan reports (100 max.) for the specified application. You can view application reports only for applications to which you have access.  \n\nPermissions required: View IQ Elements ",
		[]byte(GetReportHistoryForApplicationInputSchema),
	)
}

// GetReportHistoryForApplicationHandler is the handler function for the GetReportHistoryForApplication tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetReportHistoryForApplicationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/reports/applications/{applicationId}/history", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetReportHistoryForApplication")
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
