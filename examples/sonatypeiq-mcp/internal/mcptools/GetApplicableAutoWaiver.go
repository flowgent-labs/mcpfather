package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetApplicableAutoWaiver tool
const GetApplicableAutoWaiverInputSchema = "{\n  \"properties\": {\n    \"violationId\": {\n      \"description\": \"Enter the policy violationId for which you want to obtain the applicable auto policy waiver \",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"violationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetApplicableAutoWaiver tool (Status: 200, Content-Type: application/json)
const GetApplicableAutoWaiverResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains details for applicable auto waiver for the " + "\x60" + "violationId" + "\x60" + " specified. \n\n## Response Structure\n\n- Structure (Type: object):\n  - **scopesOperatorAny** (Type: boolean):\n  - **reachability** (Type: boolean):\n  - **ownerName** (Type: string):\n  - **ownerType** (Type: string):\n  - **threatLevel** (Type: integer, int32):\n  - **createTime** (Type: string, date-time):\n  - **creatorId** (Type: string):\n  - **pathForward** (Type: boolean):\n  - **autoPolicyWaiverId** (Type: string):\n  - **creatorName** (Type: string):\n  - **ownerId** (Type: string):\n  - **publicId** (Type: string):\n"

// NewGetApplicableAutoWaiverMCPTool creates the MCP Tool instance for GetApplicableAutoWaiver
func NewGetApplicableAutoWaiverMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplicableAutoWaiver",
		"Use this method to obtain the existing auto waiver applicable to a policy violationviolation.\n\nPermissions required: View IQ Elements",
		[]byte(GetApplicableAutoWaiverInputSchema),
	)
}

// GetApplicableAutoWaiverHandler is the handler function for the GetApplicableAutoWaiver tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicableAutoWaiverHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policyViolations/{violationId}/applicableAutoWaiver", args, []string{"violationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplicableAutoWaiver")
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
