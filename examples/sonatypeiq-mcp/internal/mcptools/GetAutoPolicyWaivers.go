package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAutoPolicyWaivers tool
const GetAutoPolicyWaiversInputSchema = "{\n  \"properties\": {\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. The response will contain waivers that are within the scope specified.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAutoPolicyWaivers tool (Status: 200, Content-Type: application/json)
const GetAutoPolicyWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains waiver details for the specified ownerType and the corresponding ownerId, grouped by the autoPolicyWaiverId.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **scopesOperatorAny** (Type: boolean):\n    - **pathForward** (Type: boolean):\n    - **publicId** (Type: string):\n    - **threatLevel** (Type: integer, int32):\n    - **autoPolicyWaiverId** (Type: string):\n    - **createTime** (Type: string, date-time):\n    - **ownerId** (Type: string):\n    - **ownerName** (Type: string):\n    - **creatorId** (Type: string):\n    - **creatorName** (Type: string):\n    - **reachability** (Type: boolean):\n    - **ownerType** (Type: string):\n"

// NewGetAutoPolicyWaiversMCPTool creates the MCP Tool instance for GetAutoPolicyWaivers
func NewGetAutoPolicyWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAutoPolicyWaivers",
		"Use this method to retrieve waiver details for all auto policy waivers for the scope specified. You can specify the scope by using the parameters ownerType and ownerId.\n\nPermissions required: View IQ Elements",
		[]byte(GetAutoPolicyWaiversInputSchema),
	)
}

// GetAutoPolicyWaiversHandler is the handler function for the GetAutoPolicyWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAutoPolicyWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/autoPolicyWaivers/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAutoPolicyWaivers")
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
