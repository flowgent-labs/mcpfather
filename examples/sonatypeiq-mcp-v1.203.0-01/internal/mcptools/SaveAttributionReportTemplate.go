package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SaveAttributionReportTemplate tool
const SaveAttributionReportTemplateInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify the details for the template as:\\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "id" + "\x60" + " is the template id.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "templateName" + "\x60" + " indicates the name of the stored template.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "documentTitle" + "\x60" + " is the title that is displayed at the top of the report.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "header" + "\x60" + " is the text that will be displayed above the " + "\x60" + "documentTitle" + "\x60" + ".\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "footer" + "\x60" + " is the text that will be displayed at the bottom of the report.\\u003cli\\u003e\\u003cli\\u003e" + "\x60" + "includeTableOfContents" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if a table of contents containing links to the components and their licenses will be added to the report.\\u003cli\\u003e" + "\x60" + "includeAppendix" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if standard license text will be grouped in the report appendix.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "includeStandardLicenseTexts" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if the standard license text will be displayed for components with no license files.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "includeSonatypeSpecialLicenses" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if Sonatype Special Licenses (e.g. Generic-Copyleft-Clause, Generic-Liberal-Clause, See-License-Clause, Identity-Clause etc.) will be included in the report.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "includeInnerSource" + "\x60" + " is " + "\x60" + "true" + "\x60" + " if InnerSource components will be included in the report.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"documentTitle\": {\n          \"type\": \"string\"\n        },\n        \"footer\": {\n          \"type\": \"string\"\n        },\n        \"header\": {\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"includeAppendix\": {\n          \"type\": \"boolean\"\n        },\n        \"includeInnerSource\": {\n          \"type\": \"boolean\"\n        },\n        \"includeSonatypeSpecialLicenses\": {\n          \"type\": \"boolean\"\n        },\n        \"includeStandardLicenseTexts\": {\n          \"type\": \"boolean\"\n        },\n        \"includeTableOfContents\": {\n          \"type\": \"boolean\"\n        },\n        \"lastUpdatedAt\": {\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"templateName\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the SaveAttributionReportTemplate tool (Status: 200, Content-Type: application/json)
const SaveAttributionReportTemplateResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of the template created or updated.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **templateName** (Type: string):\n  - **documentTitle** (Type: string):\n  - **footer** (Type: string):\n  - **lastUpdatedAt** (Type: string, date-time):\n  - **includeSonatypeSpecialLicenses** (Type: boolean):\n  - **header** (Type: string):\n  - **includeInnerSource** (Type: boolean):\n  - **id** (Type: string):\n  - **includeAppendix** (Type: boolean):\n  - **includeStandardLicenseTexts** (Type: boolean):\n  - **includeTableOfContents** (Type: boolean):\n"

// NewSaveAttributionReportTemplateMCPTool creates the MCP Tool instance for SaveAttributionReportTemplate
func NewSaveAttributionReportTemplateMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SaveAttributionReportTemplate",
		"Use this method to create or update a template.\n\nPermissions required: Review Legal Obligations For Components Licenses for the root organization",
		[]byte(SaveAttributionReportTemplateInputSchema),
	)
}

// SaveAttributionReportTemplateHandler is the handler function for the SaveAttributionReportTemplate tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SaveAttributionReportTemplateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/licenseLegalMetadata/report-template", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SaveAttributionReportTemplate")
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
