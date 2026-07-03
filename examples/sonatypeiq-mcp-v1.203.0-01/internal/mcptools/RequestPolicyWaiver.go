package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the RequestPolicyWaiver tool
const RequestPolicyWaiverInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"The request JSON should contain\\u003col\\u003e\\u003cli\\u003ecomment (optional, default null) to indicate the waiver request reason\\u003c/li\\u003e\\u003cli\\u003epolicyViolationLink (link to the policy violation page in the Lifecycle UI)\\u003c/li\\u003e\\u003cli\\u003eaddWaiverLink (link to the Add Waiver page in the Lifecycle UI)\\u003c/li\\u003e\\u003c/ol\\u003e\",\n      \"properties\": {\n        \"addWaiverLink\": {\n          \"type\": \"string\"\n        },\n        \"comment\": {\n          \"type\": \"string\"\n        },\n        \"policyViolationLink\": {\n          \"type\": \"string\"\n        },\n        \"reasonId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"policyViolationId\": {\n      \"description\": \"Enter the policyViolationId for which you want to trigger the waiver request event.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"policyViolationId\"\n  ],\n  \"type\": \"object\"\n}"

// NewRequestPolicyWaiverMCPTool creates the MCP Tool instance for RequestPolicyWaiver
func NewRequestPolicyWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RequestPolicyWaiver",
		"Deprecated since IQ Server 1.192. Triggers a 'Waiver Request' webhook event. Deprecated because the webhook event is now integrated into the policy waiver request process. Please use "+"\x60"+"api/v2/policyWaiverRequests{ownerType}/policyViolation/{policyViolationId}"+"\x60"+" instead. Scheduled for removal in December 2025.",
		[]byte(RequestPolicyWaiverInputSchema),
	)
}

// RequestPolicyWaiverHandler is the handler function for the RequestPolicyWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RequestPolicyWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/policyWaivers/waiverRequests/{policyViolationId}", args, []string{"policyViolationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "RequestPolicyWaiver")
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
