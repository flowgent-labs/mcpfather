package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetComponentVersions tool
const GetComponentVersionsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Possible values: Component identifier or packageURL (pURL) identifier in the correct format. Use a-name for JavaScript components.\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        },\n        \"packageUrl\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetComponentVersions tool (Status: 200, Content-Type: application/json)
const GetComponentVersionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Known versions of the component are returned in a string array of ascending order.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: string):\n"

// NewGetComponentVersionsMCPTool creates the MCP Tool instance for GetComponentVersions
func NewGetComponentVersionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetComponentVersions",
		"Use this method to retrieve all known versions of a component.",
		[]byte(GetComponentVersionsInputSchema),
	)
}

// GetComponentVersionsHandler is the handler function for the GetComponentVersions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetComponentVersionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/components/versions", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetComponentVersions")
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
