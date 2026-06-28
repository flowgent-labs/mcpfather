package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetAllAttributionReportTemplates tool
const GetAllAttributionReportTemplatesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAllAttributionReportTemplates tool (Status: 200, Content-Type: application/json)
const GetAllAttributionReportTemplatesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the stored template details for all reports. For each template:<ul><li>" + "\x60" + "id" + "\x60" + " is the template id.</li><li>" + "\x60" + "templateName" + "\x60" + " indicates the name of the stored template.</li><li>" + "\x60" + "documentTitle" + "\x60" + " is the title that is displayed at the top of the report.</li><li>" + "\x60" + "header" + "\x60" + " is the text that will be displayed above the " + "\x60" + "documentTitle" + "\x60" + ".</li><li>" + "\x60" + "footer" + "\x60" + " is the text that will be displayed at the bottom of the report.<li><li>" + "\x60" + "includeTableOfContents" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if a table of contents containing links to the components and their licenses will be added to the report.<li>" + "\x60" + "includeAppendix" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if standard license text will be grouped in the report appendix.</li><li>" + "\x60" + "includeStandardLicenseTexts" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if the standard license text will be displayed for components with no license files.</li><li>" + "\x60" + "includeSonatypeSpecialLicenses" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if Sonatype Special Licenses (e.g. Generic-Copyleft-Clause, Generic-Liberal-Clause, See-License-Clause, Identity-Clause etc.) will be included in the report.</li><li>" + "\x60" + "lastUpdatedAt" + "\x60" + " indicates the time the template was last updated.</li><li>" + "\x60" + "includeInnerSource" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if InnerSource components will be included in the report.</li></ul>\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **lastUpdatedAt** (Type: string, date-time):\n    - **id** (Type: string):\n    - **includeAppendix** (Type: boolean):\n    - **includeSonatypeSpecialLicenses** (Type: boolean):\n    - **includeTableOfContents** (Type: boolean):\n    - **templateName** (Type: string):\n    - **documentTitle** (Type: string):\n    - **footer** (Type: string):\n    - **includeStandardLicenseTexts** (Type: boolean):\n    - **header** (Type: string):\n    - **includeInnerSource** (Type: boolean):\n"

// NewGetAllAttributionReportTemplatesMCPTool creates the MCP Tool instance for GetAllAttributionReportTemplates
func NewGetAllAttributionReportTemplatesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllAttributionReportTemplates",
		"Use this method to retrieve templates for all reports.\n\nPermissions required: Review Legal Obligations For Components Licenses for the root organization",
		[]byte(GetAllAttributionReportTemplatesInputSchema),
	)
}

// GetAllAttributionReportTemplatesHandler is the handler function for the GetAllAttributionReportTemplates tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllAttributionReportTemplatesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/licenseLegalMetadata/report-template", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAllAttributionReportTemplates")
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
