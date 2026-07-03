package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the ScanComponents tool
const ScanComponentsInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the application internal id. Use the Applications REST API to retrieve theapplication internal id.\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"Select the request header content-type from the dropdown, depending on whether the SBOM is in XML or JSON format.\\n\\nEmbed the contents of the SBOM here or enter the file path for the SBOM. A component in the SBOM can be identified by: \\u003col\\u003e\\u003cli\\u003epackageUrl\\u003c/li\\u003e\\u003cli\\u003eComponent hash\\u003c/li\\u003e\\u003cli\\u003eComponent name and version\\u003c/li\\u003e\\u003c/ol\\u003e\\n\\nAny SPE and SWID tags for the component will be preserved in the evaluation report.\",\n      \"oneOf\": [\n        {\n          \"title\": \"Schema for application/xml\",\n          \"type\": \"string\"\n        },\n        {\n          \"title\": \"Schema for application/json\",\n          \"type\": \"string\"\n        }\n      ]\n    },\n    \"source\": {\n      \"description\": \"Specify the specification name of the SBOM to be evaluated. Allowed values are " + "\x60" + "cyclonedx" + "\x60" + " and " + "\x60" + "spdx" + "\x60" + "\",\n      \"type\": \"string\"\n    },\n    \"stageId\": {\n      \"default\": \"build\",\n      \"description\": \"Enter the stageId to run the evaluation for. The policy actions will be determined by the stage selected. Allowed values are " + "\x60" + "develop" + "\x60" + ", " + "\x60" + "build" + "\x60" + ", " + "\x60" + "stage-release" + "\x60" + ", " + "\x60" + "release" + "\x60" + " and " + "\x60" + "operate" + "\x60" + "\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"source\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ScanComponents tool (Status: 202, Content-Type: application/json)
const ScanComponentsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 202\n\n**Content-Type:** application/json\n\n> The response contains a " + "\x60" + "statusUrl" + "\x60" + " containing the applicationId and statusId that can be used to check the progress of the SBOM evaluation.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **statusUrl** (Type: string):\n"

// NewScanComponentsMCPTool creates the MCP Tool instance for ScanComponents
func NewScanComponentsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ScanComponents",
		"Use this method to perform an analysis of an SBOM.\n\nPermissions required: Evaluate Applications",
		[]byte(ScanComponentsInputSchema),
	)
}

// ScanComponentsHandler is the handler function for the ScanComponents tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ScanComponentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/scan/applications/{applicationId}/sources/{source}", args, []string{"applicationId", "source"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ScanComponents")
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
