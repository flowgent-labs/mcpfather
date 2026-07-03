package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetOpenApi tool
const GetOpenApiInputSchema = "{\n  \"properties\": {\n    \"apiType\": {\n      \"description\": \"Select the type of the API.\\u003cul\\u003e\\u003cli\\u003e " + "\x60" + "public" + "\x60" + " APIs are Generally Available and fully supported by Sonatype.\\u003c/li\\u003e\\u003cli\\u003e " + "\x60" + "experimental" + "\x60" + " APIs are not production ready, may change, and are not intended to be used in critical workloads.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"enum\": [\n        \"public\",\n        \"experimental\"\n      ],\n      \"pattern\": \"public|experimental\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"apiType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetOpenApi tool (Status: 200, Content-Type: application/json)
const GetOpenApiResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the OpenAPI documentation.\n\n## Response Structure\n\n- Structure (Type: string):\n"

// NewGetOpenApiMCPTool creates the MCP Tool instance for GetOpenApi
func NewGetOpenApiMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetOpenApi",
		"Use this method to retrieve the OpenAPI documentation for the specified type of IQ Server REST API.",
		[]byte(GetOpenApiInputSchema),
	)
}

// GetOpenApiHandler is the handler function for the GetOpenApi tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetOpenApiHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/endpoints/{apiType}", args, []string{"apiType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetOpenApi")
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
