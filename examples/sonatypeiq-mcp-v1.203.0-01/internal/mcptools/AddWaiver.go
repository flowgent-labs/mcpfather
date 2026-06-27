package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddWaiver tool
const AddWaiverInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON can include the fields\\u003col\\u003e\\u003cli\\u003eexpiryTime (default null): Sets the datetime when the waiver expires.\\u003c/li\\u003e\\u003cli\\u003ewaiverReasonId (default null): Sets the specific reason chosen for the waiver.\\u003c/li\\u003e\\u003cli\\u003ecomment (default null): Further explanation about the waiver.\\u003c/li\\u003e\\u003c/ol\\u003e\",\n      \"properties\": {\n        \"comment\": {\n          \"type\": \"string\"\n        },\n        \"expiryTime\": {\n          \"format\": \"date-time\",\n          \"type\": \"string\"\n        },\n        \"waiverReasonId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"containerImageId\": {\n      \"description\": \"Enter the container image id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"containerImageId\"\n  ],\n  \"type\": \"object\"\n}"

// NewAddWaiverMCPTool creates the MCP Tool instance for AddWaiver
func NewAddWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddWaiver",
		"Use this method to create a waiver for all policy violations of a container Image. \n\nPermissions required: Waive Policy Violations",
		[]byte(AddWaiverInputSchema),
	)
}

// AddWaiverHandler is the handler function for the AddWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/firewall/container-image/{containerImageId}/policyWaiver", args, []string{"containerImageId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddWaiver")
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
