package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetLicenseLegalComponentReport tool
const GetLicenseLegalComponentReportInputSchema = "{\n  \"properties\": {\n    \"componentIdentifier\": {\n      \"description\": \"Enter the componentIdentifier consisting of format and coordinates.\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"hash\": {\n      \"description\": \"Enter the component hash.\",\n      \"type\": \"string\"\n    },\n    \"identificationSource\": {\n      \"description\": \"Enter the identification source if a third party scan is used.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the ownerId corresponding to the ownerType.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    },\n    \"packageUrl\": {\n      \"description\": \"Enter the package URL.\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId for the report where the component was identified (required if identified by third party scan).\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetLicenseLegalComponentReport tool (Status: 200, Content-Type: application/json)
const GetLicenseLegalComponentReportResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the requested component data and the corresponding license legal metadata.\n\n1. The component data includes:<ul><li>" + "\x60" + "packageURL" + "\x60" + " is the package URL or pURL of the component.</li><li>" + "\x60" + "hash" + "\x60" + " is the truncated hash value and can be used in other REST API calls. It should not be used as a checksum.</li><li>" + "\x60" + "componentIdentifier" + "\x60" + " includes the component format and its coordinates.</li><li>" + "\x60" + "displayName" + "\x60" + " is the display name of the component.</li><li>" + "\x60" + "licenseLegalData" + "\x60" + " contains the legal data.</li><li>" + "\x60" + "stageScans" + "\x60" + " is a list and each element contains " + "\x60" + "stageName" + "\x60" + " at which the application was scanned, the " + "\x60" + "scanId" + "\x60" + " of the application scan, and the " + "\x60" + "scanDate" + "\x60" + ".</li></ul>2. The " + "\x60" + "licenseLegalMetaData" + "\x60" + " is used as a dictionary for legal data and for each license contains the license id(s), name, license text, obligations, license threat group, and whether or not it is multi license.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **licenseLegalMetadata** (Type: array):\n      - Unique Items: true\n    - **Items** (Type: object):\n      - **isMulti** (Type: boolean):\n      - **licenseId** (Type: string):\n      - **licenseName** (Type: string):\n      - **licenseText** (Type: string):\n      - **obligations** (Type: array):\n          - Unique Items: true\n        - **Items** (Type: object):\n          - **name** (Type: string):\n          - **obligationTexts** (Type: array):\n              - Unique Items: true\n            - **Items** (Type: string):\n      - **singleLicenseIds** (Type: array):\n          - Unique Items: true\n        - **Items** (Type: string):\n      - **threatGroup** (Type: object):\n        - **name** (Type: string):\n        - **threatLevel** (Type: integer, int32):\n  - **component** (Type: object):\n    - **hash** (Type: string):\n    - **licenseLegalData** (Type: object):\n      - **noticeFiles** (Type: array):\n        - **Items** (Type: object):\n          - **relPath** (Type: string):\n          - **status** (Type: string):\n              - Enum: ['enabled', 'disabled']\n          - **content** (Type: string):\n          - **id** (Type: string):\n          - **originalContentHash** (Type: string):\n      - **observedLicenses** (Type: array):\n        - **Items** (Type: string):\n      - **componentCopyrightLastUpdatedAt** (Type: string, date-time):\n      - **componentCopyrightId** (Type: string):\n      - **componentNoticesLastUpdatedByUsername** (Type: string):\n      - **componentLicensesId** (Type: string):\n      - **componentLicensesLastUpdatedAt** (Type: string, date-time):\n      - **componentNoticesLastUpdatedAt** (Type: string, date-time):\n      - **declaredLicenses** (Type: array):\n        - **Items** (Type: string):\n      - **effectiveLicenses** (Type: array):\n        - **Items** (Type: string):\n      - **licenseFiles** (Type: array):\n        - **[cyclic reference]**\n      - **sourceLinks** (Type: array):\n          - Unique Items: true\n        - **Items** (Type: object):\n          - **content** (Type: string):\n          - **id** (Type: string):\n          - **originalContent** (Type: string):\n          - **status** (Type: string):\n              - Enum: ['enabled', 'disabled']\n      - **effectiveLicenseStatus** (Type: string):\n      - **componentCopyrightLastUpdatedByUsername** (Type: string):\n      - **componentCopyrightScopeOwnerId** (Type: string):\n      - **componentLicensesLastUpdatedByUsername** (Type: string):\n      - **componentLicensesScopeOwnerId** (Type: string):\n      - **highestEffectiveLicenseThreatGroup** (Type: object):\n        - **licenseThreatGroupName** (Type: string):\n        - **licenseThreatGroupCategory** (Type: string):\n        - **licenseThreatGroupLevel** (Type: integer, int32):\n      - **componentNoticesScopeOwnerId** (Type: string):\n      - **attributions** (Type: array):\n        - **Items** (Type: object):\n          - **lastUpdatedAt** (Type: string, date-time):\n          - **lastUpdatedByUsername** (Type: string):\n          - **obligationName** (Type: string):\n          - **ownerId** (Type: string):\n          - **packageUrl** (Type: string):\n          - **componentIdentifier** (Type: object):\n            - **coordinates** (Type: object):\n              - **Additional Properties**:\n                - **property value** (Type: string):\n            - **format** (Type: string):\n          - **content** (Type: string):\n          - **id** (Type: string):\n      - **copyrights** (Type: array):\n        - **Items** (Type: object):\n          - **originalContentHash** (Type: string):\n          - **status** (Type: string):\n              - Enum: ['enabled', 'disabled']\n          - **content** (Type: string):\n          - **id** (Type: string):\n      - **obligations** (Type: array):\n        - **Items** (Type: object):\n          - **comment** (Type: string):\n          - **[cyclic reference]**\n          - **name** (Type: string):\n          - **lastUpdatedAt** (Type: string, date-time):\n          - **status** (Type: string):\n              - Enum: ['OPEN', 'IGNORED', 'FLAGGED', 'FULFILLED']\n          - **ownerId** (Type: string):\n          - **id** (Type: string):\n          - **lastUpdatedByUsername** (Type: string):\n          - **packageUrl** (Type: string):\n      - **componentNoticesId** (Type: string):\n    - **packageUrl** (Type: string):\n    - **stageScans** (Type: array):\n      - **Items** (Type: object):\n        - **scanDate** (Type: string, date-time):\n        - **scanId** (Type: string):\n        - **stageName** (Type: string):\n    - **[cyclic reference]**\n    - **displayName** (Type: string):\n"

// NewGetLicenseLegalComponentReportMCPTool creates the MCP Tool instance for GetLicenseLegalComponentReport
func NewGetLicenseLegalComponentReportMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLicenseLegalComponentReport",
		"Use this method to retrieve the raw license legal data for a component by specifying the component identifier or package URL or the component hash.\n\nPermissions required: Review Legal Obligations For Components Licenses",
		[]byte(GetLicenseLegalComponentReportInputSchema),
	)
}

// GetLicenseLegalComponentReportHandler is the handler function for the GetLicenseLegalComponentReport tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLicenseLegalComponentReportHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/licenseLegalMetadata/{ownerType}/{ownerId}/component", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLicenseLegalComponentReport")
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
