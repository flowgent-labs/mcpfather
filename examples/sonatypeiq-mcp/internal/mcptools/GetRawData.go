package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetRawData tool
const GetRawDataInputSchema = "{\n  \"properties\": {\n    \"applicationPublicId\": {\n      \"description\": \"Enter the applicationPublicId (assigned at the time of creating a new application.) \",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the reportId (scanId) created at the time of evaluating the application. application.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationPublicId\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRawData tool (Status: 200, Content-Type: application/json)
const GetRawDataResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response fields contain the 'raw' data for the reportId (scanId) specified in the method call. The fields corresponding to 'dependencyData' will indicate if the component is a direct dependency (true/false), an InnerSource component(true/false), the associated parentComponentPurls (package URLs of the parent component ownerApplicationName (name of the owner application), ownerApplicatonId (internal ID of the owner application, innerSourceComponentPurl (the package URL of the InnerSourceComponent.)\n\n## Response Structure\n\n- Structure (Type: object):\n  - **matchSummary** (Type: object):\n    - **knownComponentCount** (Type: integer, int32):\n    - **totalComponentCount** (Type: integer, int32):\n  - **components** (Type: array):\n    - **Items** (Type: object):\n      - **swid** (Type: object):\n        - **name** (Type: string):\n        - **patch** (Type: boolean):\n        - **tagId** (Type: string):\n        - **tagVersion** (Type: integer, int32):\n        - **text** (Type: object):\n          - **encoding** (Type: string):\n          - **content** (Type: string):\n          - **contentType** (Type: string):\n        - **version** (Type: string):\n      - **securityData** (Type: object):\n        - **securityIssues** (Type: array):\n          - **Items** (Type: object):\n            - **analysis** (Type: object):\n              - **detail** (Type: string):\n              - **justification** (Type: string):\n              - **response** (Type: string):\n              - **state** (Type: string):\n            - **source** (Type: string):\n            - **threatCategory** (Type: string):\n            - **url** (Type: string):\n            - **status** (Type: string):\n            - **cwe** (Type: string):\n            - **severity** (Type: number, float):\n            - **cvssVector** (Type: string):\n            - **cvssVectorSource** (Type: string):\n            - **reference** (Type: string):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **pathnames** (Type: array):\n        - **Items** (Type: string):\n      - **displayName** (Type: string):\n      - **identificationSource** (Type: string):\n      - **packageUrl** (Type: string):\n      - **dependencyData** (Type: object):\n        - **parentComponentPurls** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: string):\n        - **directDependency** (Type: boolean):\n        - **innerSource** (Type: boolean):\n        - **innerSourceData** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: object):\n            - **innerSourceComponentPurl** (Type: string):\n            - **ownerApplicationId** (Type: string):\n            - **ownerApplicationName** (Type: string):\n      - **cpe** (Type: string):\n      - **aiModelData** (Type: object):\n        - **contentTypes** (Type: array):\n          - **Items** (Type: object):\n            - **name** (Type: string):\n            - **id** (Type: string):\n        - **[cyclic reference]**\n        - **derivedFromSimilarityScore** (Type: number, double):\n      - **sha256** (Type: string):\n      - **originalPurl** (Type: string):\n      - **proprietary** (Type: boolean):\n      - **hash** (Type: string):\n      - **filenames** (Type: array):\n        - **Items** (Type: string):\n      - **thirdParty** (Type: boolean):\n      - **licenseData** (Type: object):\n        - **declaredLicenses** (Type: array):\n          - **Items** (Type: object):\n            - **licenseId** (Type: string):\n            - **licenseName** (Type: string):\n        - **effectiveLicenseThreats** (Type: array):\n          - **Items** (Type: object):\n            - **licenseThreatGroupCategory** (Type: string):\n            - **licenseThreatGroupLevel** (Type: integer, int32):\n            - **licenseThreatGroupName** (Type: string):\n        - **effectiveLicenses** (Type: array):\n          - **[cyclic reference]**\n        - **observedLicenses** (Type: array):\n          - **[cyclic reference]**\n        - **overriddenLicenses** (Type: array):\n          - **[cyclic reference]**\n        - **status** (Type: string):\n      - **matchState** (Type: string):\n  - **globalInformation** (Type: object):\n    - **dataVersionDate** (Type: string):\n"

// NewGetRawDataMCPTool creates the MCP Tool instance for GetRawData
func NewGetRawDataMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRawData",
		"Use this method to retrieve the 'raw' data generated as a result of an application evaluation. 'raw' data includes: the components identified in the application, and the licenses and vulnerabilities associated with the identified components./n/nPermissions required: View IQ Elements",
		[]byte(GetRawDataInputSchema),
	)
}

// GetRawDataHandler is the handler function for the GetRawData tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRawDataHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications/{applicationPublicId}/reports/{scanId}/raw", args, []string{"applicationPublicId", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRawData")
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
