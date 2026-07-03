package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetTransitivePolicyViolationsByOwnerStageComponent tool
const GetTransitivePolicyViolationsByOwnerStageComponentInputSchema = "{\n  \"properties\": {\n    \"componentIdentifier\": {\n      \"description\": \"Enter the component identifier and the coordinates of the component for which you want to obtain the transitive violations. This is optional, not required if package URL or hash value is provided.\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"hash\": {\n      \"description\": \"Enter the hash value of the component. This is optional, not required if component identifier or package URL is provided.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Possible values are applicationId, organizationId\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Possible values are 'application' or 'organization'\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    },\n    \"packageUrl\": {\n      \"description\": \"Enter the package URL of the component. This is optional, not required if component identifier or hash value is provided.\",\n      \"type\": \"string\"\n    },\n    \"stageId\": {\n      \"description\": \"Possible values are 'develop', 'source', 'build', 'stage-release', 'release', and, 'operate'.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\",\n    \"stageId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetTransitivePolicyViolationsByOwnerStageComponent tool (Status: 204, Content-Type: application/json)
const GetTransitivePolicyViolationsByOwnerStageComponentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 204\n\n**Content-Type:** application/json\n\n> The response contains all transitive violations detected for the component specified. In addition to the policy violation details like the name/id of the policy violated, threat level threat category, etc. the response also indicates if the violation is due to an 'InnerSource' component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **displayName** (Type: string):\n  - **hash** (Type: string):\n  - **isInnerSource** (Type: boolean):\n  - **packageUrl** (Type: string):\n  - **transitivePolicyViolations** (Type: array):\n    - **Items** (Type: object):\n      - **hash** (Type: string):\n      - **packageUrl** (Type: string):\n      - **policyId** (Type: string):\n      - **policyName** (Type: string):\n      - **threatCategory** (Type: string):\n      - **action** (Type: string):\n      - **displayName** (Type: string):\n      - **policyViolationId** (Type: string):\n      - **threatLevel** (Type: integer, int32):\n      - **[cyclic reference]**\n"

// NewGetTransitivePolicyViolationsByOwnerStageComponentMCPTool creates the MCP Tool instance for GetTransitivePolicyViolationsByOwnerStageComponent
func NewGetTransitivePolicyViolationsByOwnerStageComponentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetTransitivePolicyViolationsByOwnerStageComponent",
		"Use this method to obtain all transitive policy violations for a given component in  a specific stage. Transitive policy violations are violations caused by transitive dependencies.\n\nPermissions required: View IQ Elements",
		[]byte(GetTransitivePolicyViolationsByOwnerStageComponentInputSchema),
	)
}

// GetTransitivePolicyViolationsByOwnerStageComponentHandler is the handler function for the GetTransitivePolicyViolationsByOwnerStageComponent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetTransitivePolicyViolationsByOwnerStageComponentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyViolations/transitive/{ownerType}/{ownerId}/stages/{stageId}", args, []string{"ownerId", "ownerType", "stageId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetTransitivePolicyViolationsByOwnerStageComponent")
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
