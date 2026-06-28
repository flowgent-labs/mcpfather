package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the Set tool
const SetInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify the hash (required), comment (optional), createTime (optional), and the component identifier/package URL (required) with non-null/non-empty format and coordinates,  for the component to be claimed.\",\n      \"properties\": {\n        \"claimerId\": {\n          \"type\": \"string\"\n        },\n        \"claimerName\": {\n          \"type\": \"string\"\n        },\n        \"comment\": {\n          \"type\": \"string\"\n        },\n        \"componentIdentifier\": {\n          \"properties\": {\n            \"coordinates\": {\n              \"additionalProperties\": {\n                \"type\": \"string\"\n              },\n              \"type\": \"object\"\n            },\n            \"format\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"createTime\": {\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"hash\": {\n          \"type\": \"string\"\n        },\n        \"packageUrl\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Set tool (Status: 200, Content-Type: application/json)
const SetResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response shows the new/updated details for the claimed component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **comment** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **createTime** (Type: string, date-time):\n  - **hash** (Type: string):\n  - **packageUrl** (Type: string):\n  - **claimerId** (Type: string):\n  - **claimerName** (Type: string):\n"

// NewSetMCPTool creates the MCP Tool instance for Set
func NewSetMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Set",
		"Use this method to claim a component, or update the component details for a previously claimed component.\n\nPermissions required: Claim components",
		[]byte(SetInputSchema),
	)
}

// SetHandler is the handler function for the Set tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/claim/components", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Set")
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
