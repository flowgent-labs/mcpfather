package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the AddAutoPolicyWaiveExclusion tool
const AddAutoPolicyWaiveExclusionInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON can include the fields\\u003col\\u003e\\u003cli\\u003eapplicationPublicId\\u003c/li\\u003e\\u003cli\\u003eownerId - ID of the application or organization which will own the auto waiver exclusion\\u003c/li\\u003e\\u003cli\\u003epolicyViolationId - ID of the policy violation which the exclusion will apply to\\u003c/li\\u003e\\u003cli\\u003eautoPolicyWaiverId - ID of the auto waiver you want to apply a exclusion to\\u003c/li\\u003e\\u003cli\\u003escanId - ID of the scan which the violation being waived appeared in\\u003c/li\\u003e\\u003cli\\u003ematchStrategy (enumeration, required) can have values EXACT_COMPONENT, ALL_VERSIONS, POLICY_VIOLATION. \\u003c/li\\u003e\\u003c/ol\\u003e\",\n      \"properties\": {\n        \"applicationPublicId\": {\n          \"type\": \"string\"\n        },\n        \"autoPolicyWaiverId\": {\n          \"type\": \"string\"\n        },\n        \"matchStrategy\": {\n          \"enum\": [\n            \"EXACT_COMPONENT\",\n            \"ALL_VERSIONS\",\n            \"POLICY_VIOLATION\"\n          ],\n          \"type\": \"string\"\n        },\n        \"ownerId\": {\n          \"type\": \"string\"\n        },\n        \"policyViolationId\": {\n          \"type\": \"string\"\n        },\n        \"scanId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify which resource type owns the auto waiver you want to apply a exclusion to. Possible values are application, organization.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddAutoPolicyWaiveExclusion tool (Status: 200, Content-Type: application/json)
const AddAutoPolicyWaiveExclusionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Auto policy waiver exclusion has been created successfully.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **componentDisplayName** (Type: string):\n  - **ownerName** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n  - **componentMatchStrategy** (Type: string):\n      - Enum: ['EXACT_COMPONENT', 'ALL_VERSIONS', 'POLICY_VIOLATION']\n  - **policyViolationId** (Type: string):\n  - **vulnerabilityIdentifiers** (Type: string):\n  - **createTime** (Type: string, date-time):\n  - **autoPolicyWaiverId** (Type: string):\n  - **ownerId** (Type: string):\n  - **scanId** (Type: string):\n  - **ownerType** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **format** (Type: string):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n  - **autoPolicyWaiverExclusionId** (Type: string):\n  - **constraintFacts** (Type: array):\n    - **Items** (Type: object):\n      - **operatorName** (Type: string):\n      - **conditionFacts** (Type: array):\n        - **Items** (Type: object):\n          - **conditionIndex** (Type: integer, int32):\n          - **conditionTypeId** (Type: string):\n          - **reason** (Type: string):\n          - **reference** (Type: object):\n            - **value** (Type: string):\n            - **type** (Type: string):\n                - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n          - **summary** (Type: string):\n          - **triggerJson** (Type: string):\n      - **constraintId** (Type: string):\n      - **constraintName** (Type: string):\n  - **creatorId** (Type: string):\n  - **creatorName** (Type: string):\n  - **hash** (Type: string):\n  - **ownerPublicId** (Type: string):\n  - **policyId** (Type: string):\n  - **policyName** (Type: string):\n"

// NewAddAutoPolicyWaiveExclusionMCPTool creates the MCP Tool instance for AddAutoPolicyWaiveExclusion
func NewAddAutoPolicyWaiveExclusionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddAutoPolicyWaiveExclusion",
		"Use this method to create an auto policy waiver exclusion for a specified auto policy waiver.\n\nPermissions required: Waive Policy Violations",
		[]byte(AddAutoPolicyWaiveExclusionInputSchema),
	)
}

// AddAutoPolicyWaiveExclusionHandler is the handler function for the AddAutoPolicyWaiveExclusion tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddAutoPolicyWaiveExclusionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/autoPolicyWaiverExclusions/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddAutoPolicyWaiveExclusion")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
