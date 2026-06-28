package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetLatest tool
const GetLatestInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"description\": \"Enter the internal applicationId for the application you want to generate the SBOM. You can also retrieve the applicationId using the Application REST API.\",\n      \"type\": \"string\"\n    },\n    \"cdxVersion\": {\n      \"description\": \"Possible values are 1.1|1.2|1.3|1.4|1.5|1.6.\",\n      \"pattern\": \"1.1|1.2|1.3|1.4|1.5|1.6\",\n      \"type\": \"string\"\n    },\n    \"stageId\": {\n      \"description\": \"Enter the stageId to generate the SBOM based on the latest application policy evaluation at that stage. Allowed values for stageId are 'develop', 'source', 'build', 'stage-release', 'release', and, 'operate'.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"applicationId\",\n    \"cdxVersion\",\n    \"stageId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetLatestMCPTool creates the MCP Tool instance for GetLatest
func NewGetLatestMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLatest",
		"Use this method to generate a CycloneDX SBOM for an application.<p>Permissions Required: View IQ Elements",
		[]byte(GetLatestInputSchema),
	)
}

// GetLatestHandler is the handler function for the GetLatest tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLatestHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/cycloneDx/{cdxVersion}/{applicationId}/stages/{stageId}", args, []string{"applicationId", "cdxVersion", "stageId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLatest")
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
