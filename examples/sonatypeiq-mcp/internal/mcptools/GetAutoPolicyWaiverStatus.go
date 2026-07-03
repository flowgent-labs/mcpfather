package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAutoPolicyWaiverStatus tool
const GetAutoPolicyWaiverStatusInputSchema = "{\n  \"properties\": {\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. The response will contain status details for the active auto policy waiver, if any, that is within the scope specified.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAutoPolicyWaiverStatus tool (Status: 200, Content-Type: application/json)
const GetAutoPolicyWaiverStatusResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains auto policy waiver status details for the specified ownerType and the corresponding ownerId.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **autoPolicyWaiverOwnerName** (Type: string):\n  - **createTime** (Type: string, date-time):\n  - **hasNotReachable** (Type: boolean):\n  - **isAutoWaiverEnabled** (Type: boolean):\n  - **threatLevel** (Type: integer, int32):\n  - **autoPolicyWaiverId** (Type: string):\n  - **autoPolicyWaiverOwnerId** (Type: string):\n  - **isInherited** (Type: boolean):\n  - **scopesOperatorAny** (Type: boolean):\n  - **hasNoPathForward** (Type: boolean):\n  - **autoPolicyWaiverOwnerType** (Type: string):\n"

// NewGetAutoPolicyWaiverStatusMCPTool creates the MCP Tool instance for GetAutoPolicyWaiverStatus
func NewGetAutoPolicyWaiverStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAutoPolicyWaiverStatus",
		"Use this method to retrieve status details for any auto policy waiver enabled for the scope specified. You can specify the scope by using the parameters ownerType and ownerId.\n\nPermissions required: View IQ Elements",
		[]byte(GetAutoPolicyWaiverStatusInputSchema),
	)
}

// GetAutoPolicyWaiverStatusHandler is the handler function for the GetAutoPolicyWaiverStatus tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAutoPolicyWaiverStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/autoPolicyWaivers/{ownerType}/{ownerId}/status", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAutoPolicyWaiverStatus")
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
