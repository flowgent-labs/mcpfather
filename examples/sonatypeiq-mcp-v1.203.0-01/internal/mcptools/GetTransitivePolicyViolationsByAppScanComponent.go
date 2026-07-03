package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetTransitivePolicyViolationsByAppScanComponent tool
const GetTransitivePolicyViolationsByAppScanComponentInputSchema = "{\n  \"properties\": {\n    \"componentIdentifier\": {\n      \"description\": \"Enter the component identifier and the coordinates of the component for which you want to retrieve the transitive policy violations. This is optional, not required if package URL or hash value is provided.\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"hash\": {\n      \"description\": \"Enter the hash value for the component for which you want to retrieve the transitive policy violations in the specific scan.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the identifier for the scope specified above. E.g. applicationId\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the scope for this violation. Possible values are 'application'\",\n      \"enum\": [\n        \"application\"\n      ],\n      \"pattern\": \"application\",\n      \"type\": \"string\"\n    },\n    \"packageUrl\": {\n      \"description\": \"Enter the package URL for the component for which you want to retrieve the transitive policy violations in the specific scan.\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId/reportId corresponding to the scan.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetTransitivePolicyViolationsByAppScanComponent tool (Status: 200, Content-Type: application/json)
const GetTransitivePolicyViolationsByAppScanComponentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains violation details for all transitive violations occurring in the scan specified. The response also indicates if the violation is due to an 'InnerSource' component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **displayName** (Type: string):\n  - **hash** (Type: string):\n  - **isInnerSource** (Type: boolean):\n  - **packageUrl** (Type: string):\n  - **transitivePolicyViolations** (Type: array):\n    - **Items** (Type: object):\n      - **packageUrl** (Type: string):\n      - **policyName** (Type: string):\n      - **policyViolationId** (Type: string):\n      - **threatLevel** (Type: integer, int32):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **policyId** (Type: string):\n      - **displayName** (Type: string):\n      - **hash** (Type: string):\n      - **action** (Type: string):\n      - **threatCategory** (Type: string):\n  - **[cyclic reference]**\n"

// NewGetTransitivePolicyViolationsByAppScanComponentMCPTool creates the MCP Tool instance for GetTransitivePolicyViolationsByAppScanComponent
func NewGetTransitivePolicyViolationsByAppScanComponentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetTransitivePolicyViolationsByAppScanComponent",
		"Use this method to retrieve transitive policy violations for a given component in a specific scan.\n\nPermissions required: View IQ Elements",
		[]byte(GetTransitivePolicyViolationsByAppScanComponentInputSchema),
	)
}

// GetTransitivePolicyViolationsByAppScanComponentHandler is the handler function for the GetTransitivePolicyViolationsByAppScanComponent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetTransitivePolicyViolationsByAppScanComponentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyViolations/transitive/{ownerType}/{ownerId}/{scanId}", args, []string{"ownerId", "ownerType", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetTransitivePolicyViolationsByAppScanComponent")
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
