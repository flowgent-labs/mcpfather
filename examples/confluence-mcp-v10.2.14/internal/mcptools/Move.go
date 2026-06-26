package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Move tool
const MoveInputSchema = "{\n  \"properties\": {\n    \"attachmentId\": {\n      \"description\": \"the id of the attachment to upload the new file for.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"description\": \"The id of the content the attachment is on.\",\n      \"type\": \"string\"\n    },\n    \"newContentId\": {\n      \"type\": \"string\"\n    },\n    \"newName\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"attachmentId\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Move tool (Status: 400, Content-Type: application/json)
const MoveResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  Returned if the attachment id is invalid.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Move tool (Status: 404, Content-Type: application/json)
const MoveResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n>  Returned if no attachment is found for the attachmentId.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewMoveMCPTool creates the MCP Tool instance for Move
func NewMoveMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Move",
		"Move attachment - Move an attachment to a different content entity object. \n\n**When moving the attachment**, the name of the attachment can be updated as well. \n\nIn order to protect against XSRF attacks, because this method accepts multipart/form-data, it has XSRF protection on it. This means you must submit a header of "+"\x60"+"X-Atlassian-Token: nocheck"+"\x60"+" with the request, otherwise it will be blocked. \n\nA simple example to move an attachment with id \"456\" in a container with id \"123\" to a container with \"789\": \n\n"+"\x60"+"curl -D- -u admin:admin -X POST -H \"X-Atlassian-Token: nocheck\" \"http://myhost/rest/api/content/123/child/attachment/456/move?newContentId=789\""+"\x60"+" \n\nAn example to move the same file, while also renaming it to \"my-new-name\": \n\n"+"\x60"+"curl -D- -u admin:admin -X POST -H \"X-Atlassian-Token: nocheck\" \"http://myhost/rest/api/content/123/child/attachment/456/move?newContentId=789&newName=my-new-name\""+"\x60"+" \n\nThis can also be used to only rename an attachment: \n\n"+"\x60"+"curl -D- -u admin:admin -X POST -H \"X-Atlassian-Token: nocheck\" \"http://myhost/rest/api/content/123/child/attachment/456/move?newContentId=123&newName=my-new-name\""+"\x60"+"",
		[]byte(MoveInputSchema),
	)
}

// MoveHandler is the handler function for the Move tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func MoveHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/content/{id}/child/attachment/{attachmentId}/move", args, []string{"attachmentId", "id"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Move"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
