package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetByScanId tool
const GetByScanIdInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the applicationId for the application you want to generate the SBOM(s).\",\n      \"type\": \"string\"\n    },\n    \"format\": {\n      \"default\": \"json\",\n      \"description\": \"Enter the format for the SBOM(s) to be generated.\",\n      \"type\": \"string\"\n    },\n    \"generateCycloneDx\": {\n      \"default\": false,\n      \"description\": \"Set to " + "\x60" + "true" + "\x60" + " to generate an equivalent CycloneDx SBOM. Both SBOMs will be combined as a tar.gz archive.\",\n      \"type\": \"boolean\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId of the application scan.\",\n      \"type\": \"string\"\n    },\n    \"spdxVersion\": {\n      \"default\": \"2.3\",\n      \"description\": \"Enter the desired SPDX version, possible values are 2.2|2.3\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetByScanId tool (Status: 200, Content-Type: application/json)
const GetByScanIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The requested SBOM(s).\n\n## Response Structure\n\n- SBOM in JSON format (Type: string):\n"

// Response Template for the GetByScanId tool (Status: 200, Content-Type: application/octet-stream)
const GetByScanIdResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/octet-stream\n\n> The requested SBOM(s).\n\n## Response Structure\n\n- SBOM archive (tar.gz) (Type: string, binary):\n"

// Response Template for the GetByScanId tool (Status: 200, Content-Type: application/xml)
const GetByScanIdResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/xml\n\n> The requested SBOM(s).\n\n## Response Structure\n\n- SBOM in XML format (Type: string):\n"

// NewGetByScanIdMCPTool creates the MCP Tool instance for GetByScanId
func NewGetByScanIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetByScanId",
		"Use this method to generate SBOM(s) based on a specific application scan.\n\nPermissions required: View IQ Elemets",
		[]byte(GetByScanIdInputSchema),
	)
}

// GetByScanIdHandler is the handler function for the GetByScanId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetByScanIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/spdx/{applicationId}/reports/{scanId}", args, []string{"applicationId", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetByScanId")
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
