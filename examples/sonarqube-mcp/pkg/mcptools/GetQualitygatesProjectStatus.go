package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the GetQualitygatesProjectStatus tool
const GetQualitygatesProjectStatusInputSchema = "{\n  \"properties\": {\n    \"analysisId\": {\n      \"description\": \"Analysis id\",\n      \"type\": \"string\"\n    },\n    \"branch\": {\n      \"description\": \"Branch key\",\n      \"type\": \"string\"\n    },\n    \"projectId\": {\n      \"description\": \"Project UUID. Doesn't work with branches or pull requests\",\n      \"type\": \"string\"\n    },\n    \"projectKey\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    },\n    \"pullRequest\": {\n      \"description\": \"Pull request id\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetQualitygatesProjectStatus tool (Status: 200, Content-Type: application/json)
const GetQualitygatesProjectStatusResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Successful response\n\n## Response Structure\n\n- Response data (see SonarQube API documentation for schema details) (Type: object):\n"

// NewGetQualitygatesProjectStatusMCPTool creates the MCP Tool instance for GetQualitygatesProjectStatus
func NewGetQualitygatesProjectStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetQualitygatesProjectStatus",
		"Get the quality gate status of a project or a Compute Engine task. - Get the quality gate status of a project or a Compute Engine task. Either 'analysisId', 'projectId' or 'projectKey' must be provided The different statuses returned are: OK, WARN, ERROR, NONE. The NONE status is returned when there is no quality gate associated with the analysis. Returns an HTTP code 404 if the analysis associated with the task is not found or does not exist. Requires one of the following permissions: - 'Administer System' - 'Administer' rights on the specified project - 'Browse' on the specified project - 'Execute Analysis' on the specified project",
		[]byte(GetQualitygatesProjectStatusInputSchema),
	)
}

// GetQualitygatesProjectStatusHandler is the handler function for the GetQualitygatesProjectStatus tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetQualitygatesProjectStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/qualitygates/project_status", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetQualitygatesProjectStatus")
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
