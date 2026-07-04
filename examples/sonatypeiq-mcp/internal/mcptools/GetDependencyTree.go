package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetDependencyTree tool
const GetDependencyTreeInputSchema = "{\n  \"properties\": {\n    \"applicationPublicId\": {\n      \"description\": \"Enter the applicationPublicId created at the time of creating the application.\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \" Enter the reportId (scanId) created at the time of evaluating the application.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationPublicId\",\n    \"scanId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetDependencyTree tool (Status: 200, Content-Type: application/json)
const GetDependencyTreeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response fields contain the 'Dependency Tree' data  under the 'children' section. The 'children' section may contain more tree nodes. Every direct dependency can have zero or more transitive dependencies. Each tree node contains the packageUrl, component identifier and a dependency tree node (if it exists.) The component identifier section contains the format and coordinates for the component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **dependencyTree** (Type: object):\n    - **componentIdentifier** (Type: object):\n      - **coordinates** (Type: object):\n        - **Additional Properties**:\n          - **property value** (Type: string):\n      - **format** (Type: string):\n    - **direct** (Type: boolean):\n    - **packageUrl** (Type: string):\n    - **children** (Type: array):\n      - **[cyclic reference]**\n"

// NewGetDependencyTreeMCPTool creates the MCP Tool instance for GetDependencyTree
func NewGetDependencyTreeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetDependencyTree",
		"Use this method to retrieve the dependencies related to the component identified at the time of application evaluation. This is currently available only for Java (Maven) and NPM applications.\n\nPermissions required: View IQ Elements",
		[]byte(GetDependencyTreeInputSchema),
	)
}

// GetDependencyTreeHandler is the handler function for the GetDependencyTree tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetDependencyTreeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/applications/{applicationPublicId}/reports/{scanId}/dependencyTree", args, []string{"applicationPublicId", "scanId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetDependencyTree")
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
