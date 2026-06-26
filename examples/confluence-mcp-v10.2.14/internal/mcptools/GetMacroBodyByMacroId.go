package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetMacroBodyByMacroId tool
const GetMacroBodyByMacroIdInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"  the id of the content.\",\n      \"type\": \"string\"\n    },\n    \"macroId\": {\n      \"description\": \"the macroId to find the correct macro.\",\n      \"type\": \"string\"\n    },\n    \"version\": {\n      \"description\": \"the version of the content which the hash belongs.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"macroId\",\n    \"version\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetMacroBodyByMacroId tool (Status: 200, Content-Type: application/json)
const GetMacroBodyByMacroIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a json representation of a macro.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetMacroBodyByMacroId tool (Status: 404, Content-Type: application/json)
const GetMacroBodyByMacroIdResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to view the content, or there is no macro matching the given id or hash.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetMacroBodyByMacroIdMCPTool creates the MCP Tool instance for GetMacroBodyByMacroId
func NewGetMacroBodyByMacroIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetMacroBodyByMacroId",
		"Get macro body by macro ID - Returns the body of a macro (in storage format) with the given id. This resource is primarily used by connect applications that require the body of macro to perform their work. \n\nWhen content is created, if no macroId is specified, then Confluence will generate a random id. The id is persisted as the content is saved and only modified by Confluence if there are conflicting IDs. \n\nTo preserve backwards compatibility this resource will also match on the hash of the macro body, even if a macroId is found. This check will become redundant as pages get macroId's generated for them and transparently propagate out to all instances.",
		[]byte(GetMacroBodyByMacroIdInputSchema),
	)
}

// GetMacroBodyByMacroIdHandler is the handler function for the GetMacroBodyByMacroId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetMacroBodyByMacroIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/history/{version}/macro/id/{macroId}", args, []string{"id", "macroId", "version"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetMacroBodyByMacroId"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
