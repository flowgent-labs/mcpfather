package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetWaivers tool
const GetWaiversInputSchema = "{\n  \"properties\": {\n    \"page\": {\n      \"format\": \"int32\",\n      \"minimum\": 1,\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"format\": \"int32\",\n      \"maximum\": 100,\n      \"minimum\": 1,\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetWaivers tool (Status: 200, Content-Type: application/json)
const GetWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Policy waivers for container images.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **results**: List of items for the current page (Type: array):\n    - **Items**: List of items for the current page (Type: object):\n      - **applicationScope** (Type: string):\n      - **createTime** (Type: string, date-time):\n      - **expiryTime** (Type: string, date-time):\n      - **maxThreatLevel** (Type: integer, int32):\n      - **ownerId** (Type: string):\n      - **policyWaiverId** (Type: string):\n      - **uniqueComponentCount** (Type: integer, int64):\n      - **uniquePolicyCount** (Type: integer, int64):\n  - **total**: Total number of items (Type: integer, int64):\n  - **page**: Current page number (Type: integer, int32):\n  - **pageCount**: Total number of pages (Type: integer, int64):\n  - **pageSize**: Number of items per page (Type: integer, int32):\n"

// NewGetWaiversMCPTool creates the MCP Tool instance for GetWaivers
func NewGetWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetWaivers",
		"Use this method to get all policy waivers for container images. \n\nPermissions required: View IQ Elements",
		[]byte(GetWaiversInputSchema),
	)
}

// GetWaiversHandler is the handler function for the GetWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/container-image/policyWaiver", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetWaivers")
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
