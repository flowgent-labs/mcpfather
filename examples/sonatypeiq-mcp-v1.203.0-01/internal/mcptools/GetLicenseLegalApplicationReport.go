package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetLicenseLegalApplicationReport tool
const GetLicenseLegalApplicationReportInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the application id or public id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetLicenseLegalApplicationReport tool (Status: 200, Content-Type: application/json)
const GetLicenseLegalApplicationReportResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a list of all components in the application and the corresponding license legal metadata. For each component, the response includes component data and license legal metadata.\n\n1. The component data includes:<ul><li>" + "\x60" + "packageURL" + "\x60" + " is the package URL or pURL of the component.</li><li>" + "\x60" + "hash" + "\x60" + " is the truncated hash value and can be used in other REST API calls. It should not be used as a checksum.</li><li>" + "\x60" + "componentIdentifier" + "\x60" + " includes the component format and its coordinates.</li><li>" + "\x60" + "displayName" + "\x60" + " is the display name of the component.</li><li>" + "\x60" + "licenseLegalData" + "\x60" + " contains the legal data.</li><li>" + "\x60" + "stageScans" + "\x60" + " is a list and each element contains " + "\x60" + "stageName" + "\x60" + " at which the application was scanned, the " + "\x60" + "scanId" + "\x60" + " of the application scan, and the " + "\x60" + "scanDate" + "\x60" + ".</li></ul>2. The " + "\x60" + "licenseLegalMetaData" + "\x60" + " is used as a dictionary for legal data and for each license contains the license id(s), name, license text, obligations, license threat group, and whether or not it is multi license.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **components** (Type: array):\n    - **Items** (Type: object):\n      - **packageUrl** (Type: string):\n      - **stageScans** (Type: array):\n        - **Items** (Type: object):\n          - **stageName** (Type: string):\n          - **scanDate** (Type: string, date-time):\n          - **scanId** (Type: string):\n      - **componentIdentifier** (Type: object):\n        - **coordinates** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **format** (Type: string):\n      - **displayName** (Type: string):\n      - **hash** (Type: string):\n      - **licenseLegalData** (Type: object):\n        - **copyrights** (Type: array):\n          - **Items** (Type: object):\n            - **status** (Type: string):\n                - Enum: ['enabled', 'disabled']\n            - **content** (Type: string):\n            - **id** (Type: string):\n            - **originalContentHash** (Type: string):\n        - **sourceLinks** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: object):\n            - **content** (Type: string):\n            - **id** (Type: string):\n            - **originalContent** (Type: string):\n            - **status** (Type: string):\n                - Enum: ['enabled', 'disabled']\n        - **componentNoticesScopeOwnerId** (Type: string):\n        - **componentLicensesLastUpdatedAt** (Type: string, date-time):\n        - **obligations** (Type: array):\n          - **Items** (Type: object):\n            - **status** (Type: string):\n                - Enum: ['OPEN', 'IGNORED', 'FLAGGED', 'FULFILLED']\n            - **ownerId** (Type: string):\n            - **name** (Type: string):\n            - **id** (Type: string):\n            - **packageUrl** (Type: string):\n            - **comment** (Type: string):\n            - **[cyclic reference]**\n            - **lastUpdatedAt** (Type: string, date-time):\n            - **lastUpdatedByUsername** (Type: string):\n        - **componentLicensesId** (Type: string):\n        - **componentNoticesId** (Type: string):\n        - **componentNoticesLastUpdatedAt** (Type: string, date-time):\n        - **componentCopyrightId** (Type: string):\n        - **componentCopyrightScopeOwnerId** (Type: string):\n        - **componentNoticesLastUpdatedByUsername** (Type: string):\n        - **licenseFiles** (Type: array):\n          - **Items** (Type: object):\n            - **id** (Type: string):\n            - **originalContentHash** (Type: string):\n            - **relPath** (Type: string):\n            - **status** (Type: string):\n                - Enum: ['enabled', 'disabled']\n            - **content** (Type: string):\n        - **componentCopyrightLastUpdatedByUsername** (Type: string):\n        - **componentLicensesLastUpdatedByUsername** (Type: string):\n        - **componentCopyrightLastUpdatedAt** (Type: string, date-time):\n        - **effectiveLicenseStatus** (Type: string):\n        - **effectiveLicenses** (Type: array):\n          - **Items** (Type: string):\n        - **highestEffectiveLicenseThreatGroup** (Type: object):\n          - **licenseThreatGroupLevel** (Type: integer, int32):\n          - **licenseThreatGroupName** (Type: string):\n          - **licenseThreatGroupCategory** (Type: string):\n        - **noticeFiles** (Type: array):\n          - **[cyclic reference]**\n        - **attributions** (Type: array):\n          - **Items** (Type: object):\n            - **id** (Type: string):\n            - **lastUpdatedAt** (Type: string, date-time):\n            - **lastUpdatedByUsername** (Type: string):\n            - **obligationName** (Type: string):\n            - **ownerId** (Type: string):\n            - **packageUrl** (Type: string):\n            - **[cyclic reference]**\n            - **content** (Type: string):\n        - **componentLicensesScopeOwnerId** (Type: string):\n        - **declaredLicenses** (Type: array):\n          - **Items** (Type: string):\n        - **observedLicenses** (Type: array):\n          - **Items** (Type: string):\n  - **licenseLegalMetadata** (Type: array):\n      - Unique Items: true\n    - **Items** (Type: object):\n      - **isMulti** (Type: boolean):\n      - **licenseId** (Type: string):\n      - **licenseName** (Type: string):\n      - **licenseText** (Type: string):\n      - **obligations** (Type: array):\n          - Unique Items: true\n        - **Items** (Type: object):\n          - **name** (Type: string):\n          - **obligationTexts** (Type: array):\n              - Unique Items: true\n            - **Items** (Type: string):\n      - **singleLicenseIds** (Type: array):\n          - Unique Items: true\n        - **Items** (Type: string):\n      - **threatGroup** (Type: object):\n        - **name** (Type: string):\n        - **threatLevel** (Type: integer, int32):\n"

// NewGetLicenseLegalApplicationReportMCPTool creates the MCP Tool instance for GetLicenseLegalApplicationReport
func NewGetLicenseLegalApplicationReportMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLicenseLegalApplicationReport",
		"Use this REST API to retrieve the raw license legal data for components in an application.\n\nPermissions required: Review Legal Obligations For Components Licenses",
		[]byte(GetLicenseLegalApplicationReportInputSchema),
	)
}

// GetLicenseLegalApplicationReportHandler is the handler function for the GetLicenseLegalApplicationReport tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLicenseLegalApplicationReportHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/licenseLegalMetadata/application/{applicationId}", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLicenseLegalApplicationReport")
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
