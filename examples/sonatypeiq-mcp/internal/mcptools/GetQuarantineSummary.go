package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetQuarantineSummary tool
const GetQuarantineSummaryInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetQuarantineSummary tool (Status: 200, Content-Type: application/json)
const GetQuarantineSummaryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "repositoryCount" + "\x60" + " is the total number of repositories.</li><li>" + "\x60" + "quarantineEnabledRepositoryCount" + "\x60" + " is the total number of repositories with quarantine  capability enabled.</li><li>" + "\x60" + "quarantinedEnabled" + "\x60" + " indicates if any repository has the quarantine capability enabled.</li><li>" + "\x60" + "totalComponentCount" + "\x60" + " is the total number of components across all repositories.</li><li>" + "\x60" + "quarantinedComponentCount" + "\x60" + " is the total number of quarantined components.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **quarantineEnabled** (Type: boolean):\n  - **quarantineEnabledRepositoryCount** (Type: integer, int64):\n  - **quarantinedComponentCount** (Type: integer, int64):\n  - **repositoryCount** (Type: integer, int64):\n  - **totalComponentCount** (Type: integer, int64):\n"

// NewGetQuarantineSummaryMCPTool creates the MCP Tool instance for GetQuarantineSummary
func NewGetQuarantineSummaryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetQuarantineSummary",
		"Use this method to request a summary of quarantined components.\n\nPermissions required: View IQ Elements",
		[]byte(GetQuarantineSummaryInputSchema),
	)
}

// GetQuarantineSummaryHandler is the handler function for the GetQuarantineSummary tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetQuarantineSummaryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/quarantine/summary", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetQuarantineSummary")
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
