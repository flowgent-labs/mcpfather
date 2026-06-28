package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddAutoPolicyWaivers tool
const AddAutoPolicyWaiversInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON can be an array that include the fields\\u003col\\u003e\\u003cli\\u003ethreatLevel\\u003c/li\\u003e\\u003cli\\u003epathForward\\u003c/li\\u003e\\u003cli\\u003ereachable\\u003c/li\\u003e\\u003c/ol\\u003e\",\n      \"items\": {\n        \"properties\": {\n          \"autoPolicyWaiverId\": {\n            \"type\": \"string\"\n          },\n          \"createTime\": {\n            \"format\": \"date-time\",\n            \"type\": \"string\"\n          },\n          \"creatorId\": {\n            \"type\": \"string\"\n          },\n          \"creatorName\": {\n            \"type\": \"string\"\n          },\n          \"ownerId\": {\n            \"type\": \"string\"\n          },\n          \"ownerName\": {\n            \"type\": \"string\"\n          },\n          \"ownerType\": {\n            \"type\": \"string\"\n          },\n          \"pathForward\": {\n            \"type\": \"boolean\"\n          },\n          \"publicId\": {\n            \"type\": \"string\"\n          },\n          \"reachability\": {\n            \"type\": \"boolean\"\n          },\n          \"scopesOperatorAny\": {\n            \"type\": \"boolean\"\n          },\n          \"threatLevel\": {\n            \"format\": \"int32\",\n            \"type\": \"integer\"\n          }\n        },\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType to specify the scope. The response will contain the details for waivers within the scope.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddAutoPolicyWaivers tool (Status: 200, Content-Type: application/json)
const AddAutoPolicyWaiversResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Auto waiver has been created successfully.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **creatorId** (Type: string):\n    - **ownerId** (Type: string):\n    - **ownerType** (Type: string):\n    - **publicId** (Type: string):\n    - **autoPolicyWaiverId** (Type: string):\n    - **ownerName** (Type: string):\n    - **pathForward** (Type: boolean):\n    - **scopesOperatorAny** (Type: boolean):\n    - **threatLevel** (Type: integer, int32):\n    - **createTime** (Type: string, date-time):\n    - **creatorName** (Type: string):\n    - **reachability** (Type: boolean):\n"

// NewAddAutoPolicyWaiversMCPTool creates the MCP Tool instance for AddAutoPolicyWaivers
func NewAddAutoPolicyWaiversMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddAutoPolicyWaivers",
		"Use this method to create an auto policy waiver configuration. Only three configurations can  exist at a time for a given application or organization. With different combinations for reachable/pathForward\n\nPermissions required: Waive Policy Violations",
		[]byte(AddAutoPolicyWaiversInputSchema),
	)
}

// AddAutoPolicyWaiversHandler is the handler function for the AddAutoPolicyWaivers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddAutoPolicyWaiversHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/autoPolicyWaivers/v2/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddAutoPolicyWaivers")
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
