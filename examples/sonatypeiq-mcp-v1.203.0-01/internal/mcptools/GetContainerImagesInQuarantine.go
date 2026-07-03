package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetContainerImagesInQuarantine tool
const GetContainerImagesInQuarantineInputSchema = "{\n  \"properties\": {\n    \"page\": {\n      \"format\": \"int32\",\n      \"minimum\": 1,\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"format\": \"int32\",\n      \"maximum\": 100,\n      \"minimum\": 1,\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetContainerImagesInQuarantine tool (Status: 200, Content-Type: application/json)
const GetContainerImagesInQuarantineResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Container images in quarantine.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **page**: Current page number (Type: integer, int32):\n  - **pageCount**: Total number of pages (Type: integer, int64):\n  - **pageSize**: Number of items per page (Type: integer, int32):\n  - **results**: List of items for the current page (Type: array):\n    - **Items**: List of items for the current page (Type: object):\n      - **openTime** (Type: string, date-time):\n      - **policyViolationCount** (Type: integer, int64):\n      - **repositoryId** (Type: string):\n      - **scanId** (Type: string):\n      - **applicationId** (Type: string):\n      - **applicationPublicId** (Type: string):\n      - **applicationName** (Type: string):\n      - **repositoryPublicId** (Type: string):\n      - **threatLevel** (Type: integer, int32):\n  - **total**: Total number of items (Type: integer, int64):\n"

// NewGetContainerImagesInQuarantineMCPTool creates the MCP Tool instance for GetContainerImagesInQuarantine
func NewGetContainerImagesInQuarantineMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetContainerImagesInQuarantine",
		"Use this method to find all container images currently in quarantine.\n\nPermissions required: Read",
		[]byte(GetContainerImagesInQuarantineInputSchema),
	)
}

// GetContainerImagesInQuarantineHandler is the handler function for the GetContainerImagesInQuarantine tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetContainerImagesInQuarantineHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/firewall/container-image/policyViolations/quarantined", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetContainerImagesInQuarantine")
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
