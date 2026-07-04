package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetLabels tool
const GetLabelsInputSchema = "{\n  \"properties\": {\n    \"inherit\": {\n      \"default\": false,\n      \"description\": \"Set to " + "\x60" + "true" + "\x60" + " to retrieve inherited component labels.\",\n      \"type\": \"boolean\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the id of the application, organization or the repository.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the " + "\x60" + "ownerType" + "\x60" + " for which you want to retrieve the component label information.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetLabels tool (Status: 200, Content-Type: application/json)
const GetLabelsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains an array of component label descriptions for the application, organization or repository, as selected in the request. Each label description contains:<ul><li>" + "\x60" + "id" + "\x60" + " is the internal identifier assigned to the label.</li><li>" + "\x60" + "label" + "\x60" + " is the identifying name of the label, for e.g. 'Architecture-Deprecated'.</li><li>" + "\x60" + "description" + "\x60" + " is additional information describing the label.</li><li>" + "\x60" + "color" + "\x60" + " is the color assigned to the component label.</li><li>" + "\x60" + "ownerId" + "\x60" + " is the identifier for the ownerType selected in the request.</li><li>" + "\x60" + "ownerType" + "\x60" + " indicates if the label is for the application, organization or repository,  as selected in the request.</li></ul>If the request parameter " + "\x60" + "inherit" + "\x60" + " is set to " + "\x60" + "true" + "\x60" + " the response contains a description of component labels that are inherited from the parent. The inherited labels can be identified by the value of " + "\x60" + "ownerId" + "\x60" + " and " + "\x60" + "ownerType" + "\x60" + ".\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **label** (Type: string):\n    - **ownerId** (Type: string):\n    - **ownerType** (Type: string):\n    - **color** (Type: string):\n    - **description** (Type: string):\n    - **id** (Type: string):\n"

// NewGetLabelsMCPTool creates the MCP Tool instance for GetLabels
func NewGetLabelsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetLabels",
		"Use this method to retrieve the details for component labels for an application, organization or a repository.\n\nPermissions required: View IQ Elements",
		[]byte(GetLabelsInputSchema),
	)
}

// GetLabelsHandler is the handler function for the GetLabels tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetLabelsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/labels/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetLabels")
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
