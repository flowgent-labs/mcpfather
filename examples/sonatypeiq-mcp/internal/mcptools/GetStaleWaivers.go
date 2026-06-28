package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetStaleWaivers tool
const GetStaleWaiversInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetStaleWaivers tool (Status: 200, Content-Type: application/json)
const GetStaleWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains waiverId of the stale waiver, policyId and policyName of the policy being waived, comment, waiver scope, time created, expiry time and the waiver creator details. The response field staleEvaluations contains a list of applications or repositories that have not been evaluated since the waiver was created. \n\n## Response Structure\n\n- Structure (Type: object):\n  - **staleWaivers** (Type: array):\n    - **Items** (Type: object):\n      - **policyId** (Type: string):\n      - **creatorName** (Type: string):\n      - **policyName** (Type: string):\n      - **reasonText** (Type: string):\n      - **staleEvaluations** (Type: object):\n        - **applications** (Type: array):\n          - **Items** (Type: object):\n            - **application** (Type: object):\n              - **organizationId** (Type: string):\n              - **publicId** (Type: string):\n              - **contactUserName** (Type: string):\n              - **id** (Type: string):\n              - **name** (Type: string):\n            - **stages** (Type: array):\n              - **Items** (Type: object):\n                - **stageId** (Type: string):\n                - **lastEvaluationDate** (Type: string, date-time):\n        - **repositories** (Type: array):\n          - **Items** (Type: object):\n            - **stages** (Type: array):\n              - **[cyclic reference]**\n            - **repository** (Type: object):\n              - **type** (Type: string):\n              - **auditEnabled** (Type: boolean):\n              - **format** (Type: string):\n              - **namespaceConfusionProtectionEnabled** (Type: boolean):\n              - **policyCompliantComponentSelectionEnabled** (Type: boolean):\n              - **publicId** (Type: string):\n              - **quarantineEnabled** (Type: boolean):\n              - **repositoryId** (Type: string):\n      - **expiryTime** (Type: string, date-time):\n      - **policyWaiverReasonId** (Type: string):\n      - **scopeOwnerType** (Type: string):\n      - **comment** (Type: string):\n      - **scopeOwnerName** (Type: string):\n      - **waiverId** (Type: string):\n      - **constraintFacts** (Type: array):\n        - **Items** (Type: object):\n          - **constraintId** (Type: string):\n          - **constraintName** (Type: string):\n          - **reasons** (Type: array):\n            - **Items** (Type: object):\n              - **reason** (Type: string):\n      - **creatorId** (Type: string):\n      - **scopeOwnerId** (Type: string):\n      - **createTime** (Type: string, date-time):\n"

// NewGetStaleWaiversMCPTool creates the MCP Tool instance for GetStaleWaivers
func NewGetStaleWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetStaleWaivers",
		"Stale waivers pose a risk because they could be applied unintentionally. Use this method to retrieve stale waivers to eliminate this risk for future application evaluations. \n\nPermissions required: View IQ Elements. You can view stale waivers only for applications/repositories  to which you have access. ",
		[]byte(GetStaleWaiversInputSchema),
	)
}

// GetStaleWaiversHandler is the handler function for the GetStaleWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetStaleWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/reports/waivers/stale", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetStaleWaivers")
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
