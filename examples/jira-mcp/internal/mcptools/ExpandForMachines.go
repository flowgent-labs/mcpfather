package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the ExpandForMachines tool
const ExpandForMachinesInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"the id of the attachment to expand.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ExpandForMachines tool (Status: 200, Content-Type: application/json)
const ExpandForMachinesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> JSON representation of the attachment expanded contents. Empty entry list means that attachment cannot be expanded. It's either empty, corrupt or not an archive at all.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **totalEntryCount**: Total number of entries available (can be larger that what was asked for) (Type: integer, int32):\n      - Example: '24'\n  - **entries** (Type: array):\n      - Example: '[{\"entryIndex\":0,\"mediaType\":\"audio/mpeg\",\"name\":\"Allegro from Duet in C Major.mp3\",\"size\":1430174},{\"entryIndex\":1,\"mediaType\":\"text/rtf\",\"name\":\"lrm.rtf\",\"size\":331}]'\n    - **Items** (Type: object):\n        - Example: '[{\"entryIndex\":0,\"mediaType\":\"audio/mpeg\",\"name\":\"Allegro from Duet in C Major.mp3\",\"size\":1430174},{\"entryIndex\":1,\"mediaType\":\"text/rtf\",\"name\":\"lrm.rtf\",\"size\":331}]'\n      - **entryIndex** (Type: integer, int64):\n      - **mediaType** (Type: string):\n      - **name** (Type: string):\n      - **size** (Type: integer, int64):\n      - **abbreviatedName** (Type: string):\n"

// NewExpandForMachinesMCPTool creates the MCP Tool instance for ExpandForMachines
func NewExpandForMachinesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ExpandForMachines",
		"Get raw attachment expansion - Tries to expand an attachment. Output is raw and should be backwards-compatible through the course of time.",
		[]byte(ExpandForMachinesInputSchema),
	)
}

// ExpandForMachinesHandler is the handler function for the ExpandForMachines tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ExpandForMachinesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/attachment/{id}/expand/raw", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ExpandForMachines")
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
