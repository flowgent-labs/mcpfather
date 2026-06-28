package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddLicenseOverride tool
const AddLicenseOverrideInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Enter the license override details to add or update a license override for a component.\\nThe request body should contain the following fields:\\n - " + "\x60" + "ownerId" + "\x60" + ": Enter the id of the application, organization or the repository.\\n - " + "\x60" + "comment" + "\x60" + ": Enter a comment for the license override.\\n - " + "\x60" + "licenseIds" + "\x60" + ": Enter the license ids for the license override.\\n - " + "\x60" + "componentIdentifier" + "\x60" + ": Enter the componentIdentifier consisting of format and coordinates.\\n - " + "\x60" + "status" + "\x60" + ": Enter the status of the license override. The possible values are " + "\x60" + "OPEN" + "\x60" + ", " + "\x60" + "ACKNOWLEDGED" + "\x60" + ", " + "\x60" + "OVERRIDDEN" + "\x60" + ", " + "\x60" + "SELECTED" + "\x60" + ", and " + "\x60" + "CONFIRMED" + "\x60" + ".\",\n      \"properties\": {\n        \"comment\": {\n          \"type\": \"string\"\n        },\n        \"componentIdentifier\": {\n          \"properties\": {\n            \"coordinates\": {\n              \"additionalProperties\": {\n                \"type\": \"string\"\n              },\n              \"type\": \"object\"\n            },\n            \"format\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"licenseIds\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\",\n          \"uniqueItems\": true\n        },\n        \"ownerId\": {\n          \"type\": \"string\"\n        },\n        \"status\": {\n          \"enum\": [\n            \"OPEN\",\n            \"ACKNOWLEDGED\",\n            \"OVERRIDDEN\",\n            \"SELECTED\",\n            \"CONFIRMED\"\n          ],\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the id of the application, organization or the repository.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the " + "\x60" + "ownerType" + "\x60" + " scope for which you want to add or update a license override\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    },\n    \"where\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AddLicenseOverride tool (Status: 200, Content-Type: application/json)
const AddLicenseOverrideResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the same license override information that was added.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **ownerId** (Type: string):\n  - **status** (Type: string):\n      - Enum: ['OPEN', 'ACKNOWLEDGED', 'OVERRIDDEN', 'SELECTED', 'CONFIRMED']\n  - **comment** (Type: string):\n  - **componentIdentifier** (Type: object):\n    - **coordinates** (Type: object):\n      - **Additional Properties**:\n        - **property value** (Type: string):\n    - **format** (Type: string):\n  - **id** (Type: string):\n  - **licenseIds** (Type: array):\n      - Unique Items: true\n    - **Items** (Type: string):\n"

// NewAddLicenseOverrideMCPTool creates the MCP Tool instance for AddLicenseOverride
func NewAddLicenseOverrideMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddLicenseOverride",
		"Use this method to add or update a license override to a component for a given owner scope.\n\nPermissions required: Change Licenses",
		[]byte(AddLicenseOverrideInputSchema),
	)
}

// AddLicenseOverrideHandler is the handler function for the AddLicenseOverride tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddLicenseOverrideHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/licenseOverrides/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddLicenseOverride")
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
