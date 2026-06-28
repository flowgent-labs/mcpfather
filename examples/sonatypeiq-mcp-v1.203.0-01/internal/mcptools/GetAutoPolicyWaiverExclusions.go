package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetAutoPolicyWaiverExclusions tool
const GetAutoPolicyWaiverExclusionsInputSchema = "{\n  \"properties\": {\n    \"autoPolicyWaiverId\": {\n      \"description\": \"Enter the id of the automatic policy waiver.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the owner id.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the owner type.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    },\n    \"page\": {\n      \"default\": 1,\n      \"description\": \"Enter the page.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"default\": 10,\n      \"description\": \"Enter the page size.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"autoPolicyWaiverId\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAutoPolicyWaiverExclusions tool (Status: 200, Content-Type: application/json)
const GetAutoPolicyWaiverExclusionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successfully retrieved the auto policy waiver exclusions.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **componentIdentifier** (Type: object):\n      - **coordinates** (Type: object):\n        - **Additional Properties**:\n          - **property value** (Type: string):\n      - **format** (Type: string):\n    - **componentMatchStrategy** (Type: string):\n        - Enum: ['EXACT_COMPONENT', 'ALL_VERSIONS', 'POLICY_VIOLATION']\n    - **constraintFacts** (Type: array):\n      - **Items** (Type: object):\n        - **conditionFacts** (Type: array):\n          - **Items** (Type: object):\n            - **reason** (Type: string):\n            - **reference** (Type: object):\n              - **value** (Type: string):\n              - **type** (Type: string):\n                  - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n            - **summary** (Type: string):\n            - **triggerJson** (Type: string):\n            - **conditionIndex** (Type: integer, int32):\n            - **conditionTypeId** (Type: string):\n        - **constraintId** (Type: string):\n        - **constraintName** (Type: string):\n        - **operatorName** (Type: string):\n    - **hash** (Type: string):\n    - **ownerPublicId** (Type: string):\n    - **policyName** (Type: string):\n    - **scanId** (Type: string):\n    - **policyViolationId** (Type: string):\n    - **ownerName** (Type: string):\n    - **autoPolicyWaiverExclusionId** (Type: string):\n    - **componentDisplayName** (Type: string):\n    - **creatorId** (Type: string):\n    - **creatorName** (Type: string):\n    - **threatLevel** (Type: integer, int32):\n    - **ownerType** (Type: string):\n    - **policyId** (Type: string):\n    - **autoPolicyWaiverId** (Type: string):\n    - **createTime** (Type: string, date-time):\n    - **vulnerabilityIdentifiers** (Type: string):\n    - **ownerId** (Type: string):\n"

// NewGetAutoPolicyWaiverExclusionsMCPTool creates the MCP Tool instance for GetAutoPolicyWaiverExclusions
func NewGetAutoPolicyWaiverExclusionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAutoPolicyWaiverExclusions",
		"Use this method to retrieve auto policy waiver exclusions for the given owner and policy waiver.\n\nPermissions required: View IQ Elements",
		[]byte(GetAutoPolicyWaiverExclusionsInputSchema),
	)
}

// GetAutoPolicyWaiverExclusionsHandler is the handler function for the GetAutoPolicyWaiverExclusions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAutoPolicyWaiverExclusionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/autoPolicyWaiverExclusions/{ownerType}/{ownerId}/{autoPolicyWaiverId}", args, []string{"autoPolicyWaiverId", "ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAutoPolicyWaiverExclusions")
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
