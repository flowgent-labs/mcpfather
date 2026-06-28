package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetApplicableLabels tool
const GetApplicableLabelsInputSchema = "{\n  \"properties\": {\n    \"ownerId\": {\n      \"description\": \"Enter the id for the application, organization or repository\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the ownerType to retrieve the component label information for.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetApplicableLabels tool (Status: 200, Content-Type: application/json)
const GetApplicableLabelsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains descriptions for all component labels that are applicable to the specified owner. These include all component labels that are assigned and inherited. The response includes:<ul><li>" + "\x60" + "ownerId" + "\x60" + " is the identifier for the owner.</li><li>" + "\x60" + "ownerName" + "\x60" + " is the name for the owner.</li><li>" + "\x60" + "ownerType" + "\x60" + " indicates if the labels are for an application, organization or repository.</li> <li>" + "\x60" + "labels" + "\x60" + " is the component labels for this owner.</li></ul>Each label includes <ul><li>" + "\x60" + "id" + "\x60" + " is the internal identifier assigned to the label.</li><li>" + "\x60" + "label" + "\x60" + " is the identifying name of the label, for e.g. 'Architecture-Deprecated'.</li><li>" + "\x60" + "description" + "\x60" + " is additional information describing the label.</li><li>" + "\x60" + "color" + "\x60" + " is the color assigned to the component label.</li><li>" + "\x60" + "ownerId" + "\x60" + " is the identifier for the ownerType selected in the request.</li><li>" + "\x60" + "ownerType" + "\x60" + " indicates if the label is for the application, organization or repository, as selected in the request.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **labelsByOwner** (Type: array):\n    - **Items** (Type: object):\n      - **labels** (Type: array):\n        - **Items** (Type: object):\n          - **label** (Type: string):\n          - **ownerId** (Type: string):\n          - **ownerType** (Type: string):\n          - **color** (Type: string):\n          - **description** (Type: string):\n          - **id** (Type: string):\n      - **ownerId** (Type: string):\n      - **ownerName** (Type: string):\n      - **ownerType** (Type: string):\n          - Enum: ['application', 'organization', 'repository_container', 'repository_manager', 'repository', 'global']\n"

// NewGetApplicableLabelsMCPTool creates the MCP Tool instance for GetApplicableLabels
func NewGetApplicableLabelsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetApplicableLabels",
		"Use this method to retrieve all component labels that are applicable to the specified application, organization or repository.\n\nPermissions required: View IQ Elements",
		[]byte(GetApplicableLabelsInputSchema),
	)
}

// GetApplicableLabelsHandler is the handler function for the GetApplicableLabels tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetApplicableLabelsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/labels/{ownerType}/{ownerId}/applicable", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetApplicableLabels")
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
