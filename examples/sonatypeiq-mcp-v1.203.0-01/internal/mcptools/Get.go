package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the Get tool
const GetInputSchema = "{\n  \"properties\": {\n    \"hash\": {\n      \"description\": \"The hash of the claimed component.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"hash\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Get tool (Status: 200, Content-Type: application/json)
const GetResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the truncated SHA1 hash of the component, the datetime when the component was published (not the time it was claimed), the format and coordinates of the claimed component (componentIdentifier) and the package URL of the claimed component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **claimerId** (Type: string):\n  - **claimerName** (Type: string):\n  - **comment** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **createTime** (Type: string, date-time):\n  - **hash** (Type: string):\n  - **packageUrl** (Type: string):\n"

// NewGetMCPTool creates the MCP Tool instance for Get
func NewGetMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Get",
		"Use this method to retrieve details of a claimed component by specifying its hash.\n\nPermissions required: Claim components",
		[]byte(GetInputSchema),
	)
}

// GetHandler is the handler function for the Get tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/claim/components/{hash}", args, []string{"hash"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Get")
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
