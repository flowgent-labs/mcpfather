package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetSourceControl tool
const GetSourceControlInputSchema = "{\n  \"properties\": {\n    \"internalOwnerId\": {\n      \"description\": \"Enter the ownerId corresponding to the ownerType.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the ownerType for the pull requests.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetSourceControl tool (Status: 200, Content-Type: application/json)
const GetSourceControlResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains " + "\x60" + "results" + "\x60" + " which is a list of elements, each including: <ul><li>" + "\x60" + "startTime" + "\x60" + " indicates the start time of the pull request.</li><li>" + "\x60" + "title" + "\x60" + " indicates the title of the pull request.</li><li>" + "\x60" + "exceptionThrown" + "\x60" + " indicates if the pull request caused an exception.</li><li>" + "\x60" + "successful" + "\x60" + " indicates if the pull request was successful.</li><li>" + "\x60" + "totalTime" + "\x60" + " indicates the total time taken to complete the pull request.</li><li>" + "\x60" + "reasoning" + "\x60" + " indicates the summary of the outcome of the pull request.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **results** (Type: array):\n    - **Items** (Type: object):\n      - **reasoning** (Type: string):\n      - **startTime** (Type: string, date-time):\n      - **successful** (Type: boolean):\n      - **title** (Type: string):\n      - **totalTime** (Type: integer, int64):\n      - **exceptionThrown** (Type: boolean):\n"

// NewGetSourceControlMCPTool creates the MCP Tool instance for GetSourceControl
func NewGetSourceControlMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSourceControl",
		"Use this method to view the source control pull request metrics.\n\nPermissions required: View IQ Elements",
		[]byte(GetSourceControlInputSchema),
	)
}

// GetSourceControlHandler is the handler function for the GetSourceControl tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSourceControlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/sourceControlMetrics/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSourceControl")
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
