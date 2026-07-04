package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetComponentDetails tool
const GetComponentDetailsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"You can retrieve component data in any one of the 3 ways via:\\n1. Component identifier\\n2. Package URL\\n3. Hash\",\n      \"properties\": {\n        \"components\": {\n          \"items\": {\n            \"properties\": {\n              \"componentIdentifier\": {\n                \"properties\": {\n                  \"coordinates\": {\n                    \"additionalProperties\": {\n                      \"type\": \"string\"\n                    },\n                    \"type\": \"object\"\n                  },\n                  \"format\": {\n                    \"type\": \"string\"\n                  }\n                },\n                \"type\": \"object\"\n              },\n              \"displayName\": {\n                \"type\": \"string\"\n              },\n              \"hash\": {\n                \"type\": \"string\"\n              },\n              \"originalPurl\": {\n                \"type\": \"string\"\n              },\n              \"packageUrl\": {\n                \"type\": \"string\"\n              },\n              \"proprietary\": {\n                \"type\": \"boolean\"\n              },\n              \"sha256\": {\n                \"type\": \"string\"\n              },\n              \"thirdParty\": {\n                \"type\": \"boolean\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetComponentDetails tool (Status: 200, Content-Type: application/json)
const GetComponentDetailsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a detailed description of the component. The hash value returned here is truncated and not intended to be used as a checksum. It can be used as an identifier to pass to other REST API calls.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **componentDetails** (Type: array):\n    - **Items** (Type: object):\n      - **matchState** (Type: string):\n      - **relativePopularity** (Type: integer, int32, nullable):\n          - Nullable: true\n      - **component** (Type: object):\n        - **hash** (Type: string):\n        - **originalPurl** (Type: string):\n        - **packageUrl** (Type: string):\n        - **proprietary** (Type: boolean):\n        - **sha256** (Type: string):\n        - **thirdParty** (Type: boolean):\n        - **componentIdentifier** (Type: object):\n          - **coordinates** (Type: object):\n            - **Additional Properties**:\n              - **property value** (Type: string):\n          - **format** (Type: string):\n        - **displayName** (Type: string):\n      - **hygieneRating** (Type: string, nullable):\n          - Nullable: true\n      - **securityData** (Type: object):\n        - **securityIssues** (Type: array):\n          - **Items** (Type: object):\n            - **severity** (Type: number, float):\n            - **analysis** (Type: object):\n              - **detail** (Type: string):\n              - **justification** (Type: string):\n              - **response** (Type: string):\n              - **state** (Type: string):\n            - **status** (Type: string):\n            - **url** (Type: string):\n            - **reference** (Type: string):\n            - **cvssVectorSource** (Type: string):\n            - **source** (Type: string):\n            - **threatCategory** (Type: string):\n            - **cvssVector** (Type: string):\n            - **cwe** (Type: string):\n      - **catalogDate** (Type: string, date-time):\n      - **policyData** (Type: object):\n        - **policyViolations** (Type: array):\n          - **Items** (Type: object):\n            - **policyId** (Type: string):\n            - **threatLevel** (Type: integer, int32):\n            - **constraintViolations** (Type: array):\n              - **Items** (Type: object):\n                - **reasons** (Type: array):\n                  - **Items** (Type: object):\n                    - **reason** (Type: string):\n                    - **reference** (Type: object):\n                      - **value** (Type: string):\n                      - **type** (Type: string):\n                          - Enum: ['SECURITY_VULNERABILITY_REFID', 'SAST_FINDING_ID']\n                - **constraintId** (Type: string):\n                - **constraintName** (Type: string):\n            - **openTime** (Type: string, date-time):\n            - **policyViolationId** (Type: string):\n            - **policyName** (Type: string):\n            - **waiveTime** (Type: string, date-time):\n            - **fixTime** (Type: string, date-time):\n            - **legacyViolationTime** (Type: string, date-time):\n      - **projectData** (Type: object):\n        - **firstReleaseDate** (Type: string, date-time):\n        - **lastReleaseDate** (Type: string, date-time):\n        - **projectMetadata** (Type: object):\n          - **description** (Type: string):\n          - **organization** (Type: string):\n        - **sourceControlManagement** (Type: object):\n          - **scmDetails** (Type: object):\n            - **commitsPerMonth** (Type: integer, int32):\n            - **uniqueDevsPerMonth** (Type: integer, int32):\n          - **scmMetadata** (Type: object):\n            - **forks** (Type: integer, int32):\n            - **stars** (Type: integer, int32):\n          - **scmUrl** (Type: string):\n      - **integrityRating** (Type: string, nullable):\n          - Nullable: true\n      - **licenseData** (Type: object):\n        - **overriddenLicenses** (Type: array):\n          - **Items** (Type: object):\n            - **licenseId** (Type: string):\n            - **licenseName** (Type: string):\n        - **status** (Type: string):\n        - **declaredLicenses** (Type: array):\n          - **[cyclic reference]**\n        - **effectiveLicenses** (Type: array):\n          - **[cyclic reference]**\n        - **observedLicenses** (Type: array):\n          - **[cyclic reference]**\n"

// NewGetComponentDetailsMCPTool creates the MCP Tool instance for GetComponentDetails
func NewGetComponentDetailsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetComponentDetails",
		"Use this method to retrieve data related to a component.",
		[]byte(GetComponentDetailsInputSchema),
	)
}

// GetComponentDetailsHandler is the handler function for the GetComponentDetails tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetComponentDetailsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/components/details", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetComponentDetails")
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
