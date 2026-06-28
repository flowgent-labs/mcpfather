package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetAttachments tool
const GetAttachmentsInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"(optional) a comma separated list of properties to expand on the Attachments returned. Optional.\",\n      \"type\": \"string\"\n    },\n    \"filename\": {\n      \"description\": \"(optional) filter parameter to return only the Attachment with the matching file name.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"The id of the content the attachment is on.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 50,\n      \"description\": \"(optional) how many items should be returned after the start index.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"mediaType\": {\n      \"description\": \"(optional) a comma separated list of properties to expand on the Attachments returned.\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"(optional) the index of the first item within the result set that should be returned.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAttachments tool (Status: 201, Content-Type: application/json)
const GetAttachmentsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of a list of attachment Content entities.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n"

// Response Template for the GetAttachments tool (Status: 404, Content-Type: application/json)
const GetAttachmentsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n>  Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetAttachmentsMCPTool creates the MCP Tool instance for GetAttachments
func NewGetAttachmentsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAttachments",
		"Get attachment - Returns a paginated list of attachment Content entities within a single container.",
		[]byte(GetAttachmentsInputSchema),
	)
}

// GetAttachmentsHandler is the handler function for the GetAttachments tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAttachmentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/child/attachment", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAttachments")
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
