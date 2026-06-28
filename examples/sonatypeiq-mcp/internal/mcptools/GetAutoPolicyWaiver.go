package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAutoPolicyWaiver tool
const GetAutoPolicyWaiverInputSchema = "{\n  \"properties\": {\n    \"autoPolicyWaiverId\": {\n      \"description\": \"Enter the autoPolicyWaiverId for which you want to retrieve the auto policy waiver details.\",\n      \"type\": \"string\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. The response will contain the details for waivers within the scope.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"autoPolicyWaiverId\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAutoPolicyWaiver tool (Status: 200, Content-Type: application/json)
const GetAutoPolicyWaiverResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains waiver details corresponding to the auto policy waiver id specified.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **creatorName** (Type: string):\n  - **reachability** (Type: boolean):\n  - **ownerType** (Type: string):\n  - **scopesOperatorAny** (Type: boolean):\n  - **pathForward** (Type: boolean):\n  - **publicId** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n  - **autoPolicyWaiverId** (Type: string):\n  - **createTime** (Type: string, date-time):\n  - **ownerId** (Type: string):\n  - **ownerName** (Type: string):\n  - **creatorId** (Type: string):\n"

// NewGetAutoPolicyWaiverMCPTool creates the MCP Tool instance for GetAutoPolicyWaiver
func NewGetAutoPolicyWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAutoPolicyWaiver",
		"Use this method to retrieve auto policy waiver details for the autoPolicyWaiverId specified.\n\nPermissions required: View IQ Elements",
		[]byte(GetAutoPolicyWaiverInputSchema),
	)
}

// GetAutoPolicyWaiverHandler is the handler function for the GetAutoPolicyWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAutoPolicyWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/autoPolicyWaivers/{ownerType}/{ownerId}/{autoPolicyWaiverId}", args, []string{"autoPolicyWaiverId", "ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAutoPolicyWaiver")
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
