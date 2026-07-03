package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddLabel tool
const AddLabelInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify a label name, description and color for the label. Valid values for color are " + "\x60" + "light-red" + "\x60" + " , " + "\x60" + "light-green" + "\x60" + " , " + "\x60" + "light-blue" + "\x60" + " , " + "\x60" + "light-purple" + "\x60" + ", " + "\x60" + "dark-red" + "\x60" + " , " + "\x60" + "dark-green" + "\x60" + " , " + "\x60" + "dark-blue" + "\x60" + " , " + "\x60" + "dark-purple" + "\x60" + " , " + "\x60" + "orange" + "\x60" + " , " + "\x60" + "yellow" + "\x60" + ". Do not enter value for the " + "\x60" + "id" + "\x60" + " field.\",\n      \"properties\": {\n        \"color\": {\n          \"type\": \"string\"\n        },\n        \"description\": {\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"label\": {\n          \"type\": \"string\"\n        },\n        \"ownerId\": {\n          \"type\": \"string\"\n        },\n        \"ownerType\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the id for the selected ownerType.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the ownerType to which the label will be assigned.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddLabel tool (Status: 200, Content-Type: application/json)
const AddLabelResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains label details sent in the request and the " + "\x60" + "id" + "\x60" + " for the label created.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **color** (Type: string):\n  - **description** (Type: string):\n  - **id** (Type: string):\n  - **label** (Type: string):\n  - **ownerId** (Type: string):\n  - **ownerType** (Type: string):\n"

// NewAddLabelMCPTool creates the MCP Tool instance for AddLabel
func NewAddLabelMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddLabel",
		"Use this method to create and assign a component label to an application, organization or repository.\n\nPermissions required: Edit IQ Elements",
		[]byte(AddLabelInputSchema),
	)
}

// AddLabelHandler is the handler function for the AddLabel tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddLabelHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/labels/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddLabel")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
