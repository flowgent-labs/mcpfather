package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Update2 tool
const Update2InputSchema = "{\n  \"properties\": {\n    \"asyncReconciliation\": {\n      \"default\": false,\n      \"type\": \"boolean\"\n    },\n    \"body\": {\n      \"description\": \"new content to be created.\"\n    },\n    \"conflictPolicy\": {\n      \"description\": \"the conflict policy, default value: \\u003ccode\\u003eabort\\u003ccode\\u003e\",\n      \"type\": \"string\"\n    },\n    \"contentId\": {\n      \"description\": \"  the id of the content.\",\n      \"type\": \"string\"\n    },\n    \"status\": {\n      \"description\": \"the existing status of the content to be updated.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"contentId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Update2 tool (Status: 200, Content-Type: application/json)
const Update2ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of a piece of content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update2 tool (Status: 400, Content-Type: application/json)
const Update2ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if no space or no content type, or setup a wrong version type set to content, or status param is not draft and status content is current\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Update2 tool (Status: 404, Content-Type: application/json)
const Update2ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if can not find draft with current content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdate2MCPTool creates the MCP Tool instance for Update2
func NewUpdate2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Update2",
		"Update content - Updates a piece of Content, including changes to content status. \n\nTo update a piece of content you must increment the "+"\x60"+"version.number"+"\x60"+", supplying the number of the version you are creating. The "+"\x60"+"title"+"\x60"+" property can be updated on all content, "+"\x60"+"body"+"\x60"+" can be updated on all content that has a body (not attachments). For instance to update the content of a blogpost that currently has version 1:\n\n"+"\x60"+"PUT /rest/api/content/456"+"\x60"+"\n\n"+"\x60"+""+"\x60"+""+"\x60"+"json\n{\n   \"version\":{\n       \"number\": 2\n   },\n   \"title\":\"My new title\",\n   \"type\":\"page\",\n   \"body\":{\n        \"storage\":{\n           \"value\":\"<p>New page data.</p>\",\n           \"representation\":\"storage\"\n      }\n   }\n}\n"+"\x60"+""+"\x60"+""+"\x60"+"\n\nTo update a page and change its parent page, supply the "+"\x60"+"ancestors"+"\x60"+" property with the request with the parent as the first ancestor i.e. to move a page to be a child of page with ID 789:\n\n"+"\x60"+"PUT /rest/api/content/456"+"\x60"+"\n\n"+"\x60"+""+"\x60"+""+"\x60"+"json\n{\n   \"version\":{\n       \"number\": 2\n   },\n   \"ancestors\": [{\"id\":789}],\n   \"type\":\"page\",\n   \"body\":{\n        \"storage\":{\n           \"value\":\"<p>New page data.</p>\",\n           \"representation\":\"storage\"\n      }\n   }\n}\n"+"\x60"+""+"\x60"+""+"\x60"+"\n\nChanging status\n\nTo restore a piece of content that has the status of trashed the content must have it's "+"\x60"+"version"+"\x60"+" incremented, and "+"\x60"+"status"+"\x60"+" set to "+"\x60"+"current"+"\x60"+". No other field modifications will be performed when restoring a piece of content from the trash.\n\nRequest example to restore from trash: "+"\x60"+"{\"id\": \"557059\",\"status\": \"current\",\"version\": {\"number\": 2}}"+"\x60"+"\n\nIf the content you're updating has a draft, specifying "+"\x60"+"status=draft"+"\x60"+" will delete that draft and the "+"\x60"+"body"+"\x60"+" of the content will be replaced with the "+"\x60"+"body"+"\x60"+" specified in the request.\n\nRequest example to delete a draft:\n\n"+"\x60"+"PUT:  http://localhost:9096/confluence/rest/api/content/2149384202?status=draft"+"\x60"+"\n\n"+"\x60"+""+"\x60"+""+"\x60"+"json\n{\n   \"id\":\"2149384202\",\n   \"status\":\"current\",\n   \"version\":{\n      \"number\":4\n   },\n   \"space\":{\n      \"key\":\"TST\"\n   },\n   \"type\":\"page\",\n   \"title\":\"page title\",\n   \"body\":{\n      \"storage\":{\n         \"value\":\"<p>New page data.</p>\",\n         \"representation\":\"storage\"\n      }\n   }\n}\n"+"\x60"+""+"\x60"+""+"\x60"+"\n\nChanging page position\n\nTo set page position, supply the "+"\x60"+"position"+"\x60"+" property in the request body with a positive integer. Content with unset positions will have a "+"\x60"+"position"+"\x60"+" value of -1. To unset a content position, supply "+"\x60"+"position"+"\x60"+" property with -1.\n\nRequest example to set page position to 1\n\n"+"\x60"+"PUT /rest/api/content/2149384202"+"\x60"+"\n\n"+"\x60"+""+"\x60"+""+"\x60"+"json\n{\n   \"id\":\"2149384202\",\n   \"version\":{\n      \"number\":2\n   },\n   \"type\":\"page\",\n   \"title\":\"page title\",\n   \"position\":1\n}\n"+"\x60"+""+"\x60"+""+"\x60"+"\n\n Request example to unset page position \n\n"+"\x60"+"PUT /rest/api/content/2149384202"+"\x60"+"\n\n"+"\x60"+""+"\x60"+""+"\x60"+"json\n{\n   \"id\":\"2149384202\",\n   \"version\":{\n      \"number\":2\n   },\n   \"type\":\"page\",\n   \"position\":-1\n}\n"+"\x60"+""+"\x60"+""+"\x60"+"\n\n",
		[]byte(Update2InputSchema),
	)
}

// Update2Handler is the handler function for the Update2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Update2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/content/{contentId}", args, []string{"contentId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Update2"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
