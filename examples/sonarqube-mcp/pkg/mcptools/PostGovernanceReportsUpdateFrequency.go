package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostGovernanceReportsUpdateFrequency tool
const PostGovernanceReportsUpdateFrequencyInputSchema = "{\n  \"properties\": {\n    \"branchKey\": {\n      \"description\": \"The branch key\",\n      \"type\": \"string\"\n    },\n    \"componentId\": {\n      \"description\": \"The portfolio, sub-portfolio, project or application id\",\n      \"type\": \"string\"\n    },\n    \"componentKey\": {\n      \"description\": \"The portfolio, sub-portfolio, project or application key\",\n      \"type\": \"string\"\n    },\n    \"frequency\": {\n      \"description\": \"Frequency at which the report is sent. If not set, the specific frequency for the component is deleted and the global frequency is used.\",\n      \"enum\": [\n        \"daily\",\n        \"weekly\",\n        \"monthly\"\n      ],\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewPostGovernanceReportsUpdateFrequencyMCPTool creates the MCP Tool instance for PostGovernanceReportsUpdateFrequency
func NewPostGovernanceReportsUpdateFrequencyMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostGovernanceReportsUpdateFrequency",
		"Update the frequency at which a report is sent for a portfolio, a sub-portfolio, a project or an application. - Update the frequency at which a report is sent for a portfolio, a sub-portfolio, a project or an application. Requires the 'Administer' permission.",
		[]byte(PostGovernanceReportsUpdateFrequencyInputSchema),
	)
}

// PostGovernanceReportsUpdateFrequencyHandler is the handler function for the PostGovernanceReportsUpdateFrequency tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostGovernanceReportsUpdateFrequencyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/governance_reports/update_frequency", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostGovernanceReportsUpdateFrequency")
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
