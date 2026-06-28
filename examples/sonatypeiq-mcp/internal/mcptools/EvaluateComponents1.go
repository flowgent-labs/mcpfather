package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the EvaluateComponents1 tool
const EvaluateComponents1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Provide the array of component identifiers to be evaluated. Each component requires **one of the following combinations**:\\n\\n**For Coordinate-Based Formats (golang, conan, cargo, cocoapods, cran, conda, composer, hf-model):**\\n- **packageUrl only** - Hash is optional\\n- **packageUrl + hash** - Hash is optional, will be used if provided\\n- **pathname + hash** - Hash required (pathname approach always requires hash)\\n\\n**For Hash-Based Formats (maven, npm, pypi, nuget, docker, rubygems, etc.):**\\n- **packageUrl + hash** - Hash REQUIRED to identify exact file content\\n- **pathname + hash** - Hash REQUIRED\\n- Providing packageUrl without hash for these formats will result in incorrect identification\\n\\nA maximum of 100 components can be evaluated in one request.\",\n      \"properties\": {\n        \"components\": {\n          \"items\": {\n            \"properties\": {\n              \"hash\": {\n                \"type\": \"string\"\n              },\n              \"packageUrl\": {\n                \"type\": \"string\"\n              },\n              \"pathname\": {\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"repositoryId\": {\n      \"description\": \"Enter the repository ID.\",\n      \"type\": \"string\"\n    },\n    \"repositoryManagerId\": {\n      \"description\": \"Enter the repository manager ID.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"repositoryId\",\n    \"repositoryManagerId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the EvaluateComponents1 tool (Status: 200, Content-Type: application/json)
const EvaluateComponents1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the evaluation results.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **repositoryManagerId** (Type: string):\n  - **repositoryPublicId** (Type: string):\n  - **repositoryType** (Type: string):\n  - **results** (Type: array):\n    - **Items** (Type: object):\n      - **component** (Type: object):\n        - **hash** (Type: string):\n        - **packageUrl** (Type: string):\n        - **pathname** (Type: string):\n      - **policyViolations** (Type: array):\n        - **Items** (Type: object):\n          - **threatLevel** (Type: integer, int32):\n          - **waiveTime** (Type: string, date-time):\n          - **fixTime** (Type: string, date-time):\n          - **legacyViolationTime** (Type: string, date-time):\n          - **openTime** (Type: string, date-time):\n          - **constraintViolations** (Type: array):\n            - **Items** (Type: object):\n              - **constraintId** (Type: string):\n              - **constraintName** (Type: string):\n              - **reasons** (Type: array):\n                - **Items** (Type: object):\n                  - **reason** (Type: string):\n                  - **reference** (Type: object):\n                    - **type** (Type: string):\n                        - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n                    - **value** (Type: string):\n          - **policyId** (Type: string):\n          - **policyName** (Type: string):\n          - **policyViolationId** (Type: string):\n      - **quarantineDate** (Type: string, date-time):\n      - **quarantined** (Type: boolean):\n      - **catalogDate** (Type: string, date-time):\n  - **repositoryId** (Type: string):\n"

// NewEvaluateComponents1MCPTool creates the MCP Tool instance for EvaluateComponents1
func NewEvaluateComponents1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"EvaluateComponents1",
		"Use this method to evaluate components (max. 100).\n\n**Hash Requirements by Format Type:**\n\n**Coordinate-Based Formats** (hash NOT required when using packageUrl):\n- golang, conan, cargo, cocoapods, cran, conda, composer, hf-model\n- These formats identify components by coordinates (name+version) rather than file hash\n- When using packageUrl, hash is optional\n- If hash is provided, it will be used\n\n**Hash-Based Formats** (hash REQUIRED):\n- maven, npm, pypi, nuget, docker, rubygems, and others\n- Hash must ALWAYS be provided to identify the exact file content\n- Hash is required regardless of whether pathname or packageUrl is used\n- Missing hash for these formats will result in incorrect component identification\n\nPermissions required: Evaluate Individual Components",
		[]byte(EvaluateComponents1InputSchema),
	)
}

// EvaluateComponents1Handler is the handler function for the EvaluateComponents1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func EvaluateComponents1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/firewall/components/{repositoryManagerId}/{repositoryId}/evaluate", args, []string{"repositoryId", "repositoryManagerId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "EvaluateComponents1")
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
