package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetConfiguration tool
const GetConfigurationInputSchema = "{\n  \"properties\": {\n    \"direct\": {\n      \"default\": false,\n      \"description\": \"Set to true to retrieve only direct configuration, false (default) to retrieve merged configuration from hierarchy\",\n      \"type\": \"boolean\"\n    },\n    \"ownerId\": {\n      \"description\": \"The internal ID of the owner\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"The owner type (application or organization)\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetConfiguration tool (Status: 200, Content-Type: application/json)
const GetConfigurationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:\n<ul>\n<li>" + "\x60" + "configuration" + "\x60" + " - the CI integration configuration as a JSON object</li>\n<li>" + "\x60" + "source" + "\x60" + " - a map of field names to owner IDs indicating provenance (empty for direct queries)</li>\n</ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **data** (Type: object):\n    - **failBuildOnScanningErrors** (Type: boolean):\n    - **download** (Type: object):\n      - **iqCliUrl** (Type: string):\n      - **iqCliVersion** (Type: string):\n    - **failBuildOnPolicyWarnings** (Type: boolean):\n    - **enableDebugLogging** (Type: boolean):\n    - **failBuildOnReachabilityErrors** (Type: boolean):\n    - **unstableBuildOnPolicyWarnings** (Type: boolean):\n    - **sarifFile** (Type: string):\n    - **failBuildOnNetworkError** (Type: boolean):\n    - **parameterPriority** (Type: string):\n    - **resultFile** (Type: string):\n    - **moduleExcludes** (Type: array):\n      - **Items** (Type: string):\n    - **scanPatterns** (Type: array):\n      - **Items** (Type: string):\n    - **reachability** (Type: object):\n      - **failOnError** (Type: boolean):\n      - **javaAnalysis** (Type: object):\n        - **enabled** (Type: boolean):\n        - **entrypointStrategy** (Type: string):\n        - **namespaces** (Type: array):\n          - **Items** (Type: string):\n      - **javaScriptAnalysis** (Type: object):\n        - **enabled** (Type: boolean):\n        - **jsExcludes** (Type: array):\n          - **Items** (Type: string):\n        - **jsSources** (Type: array):\n          - **Items** (Type: string):\n        - **nodeJsExecutable** (Type: string):\n        - **projectRoot** (Type: string):\n    - **proxy** (Type: object):\n      - **host** (Type: string):\n    - **advancedProperties** (Type: array):\n      - **Items** (Type: string):\n  - **source** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: string):\n"

// NewGetConfigurationMCPTool creates the MCP Tool instance for GetConfiguration
func NewGetConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetConfiguration",
		"Use this method to retrieve CI integration configuration for the specified owner.\n\nSet the "+"\x60"+"direct"+"\x60"+" query parameter to "+"\x60"+"true"+"\x60"+" to retrieve only the configuration directly associated with the specified owner. Set it to "+"\x60"+"false"+"\x60"+" (default) to retrieve the merged configuration from the organization hierarchy, where configurations from parent organizations are combined with lower levels taking precedence.\n\nThe response includes a "+"\x60"+"source"+"\x60"+" map that indicates which owner (organization or application) contributed each configuration field when using merged mode.\n\nPermissions required: View IQ Elements",
		[]byte(GetConfigurationInputSchema),
	)
}

// GetConfigurationHandler is the handler function for the GetConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/ci/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetConfiguration")
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
