package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetAppliedLicenseOverrides tool
const GetAppliedLicenseOverridesInputSchema = "{\n  \"properties\": {\n    \"componentIdentifier\": {\n      \"description\": \"Enter the componentIdentifier consisting of format and coordinates as a JSON e.g., " + "\x60" + "?componentIdentifier={\\\"format\\\":\\\"maven\\\",\\\"coordinates\\\":\\\"{...}}\\\"}\",\n      \"properties\": {\n        \"coordinates\": {\n          \"additionalProperties\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"object\"\n        },\n        \"format\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"ownerId\": {\n      \"description\": \"Enter the id of the application, organization or the repository.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the " + "\x60" + "ownerType" + "\x60" + " for which you want to retrieve the applied license overrides.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_container\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository|repository_manager|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"componentIdentifier\",\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAppliedLicenseOverrides tool (Status: 200, Content-Type: application/json)
const GetAppliedLicenseOverridesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the license overrides for the component.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **licenseOverridesByOwner** (Type: array):\n    - **Items** (Type: object):\n      - **licenseOverride** (Type: object):\n        - **licenseIds** (Type: array):\n            - Unique Items: true\n          - **Items** (Type: string):\n        - **ownerId** (Type: string):\n        - **status** (Type: string):\n            - Enum: ['OPEN', 'ACKNOWLEDGED', 'OVERRIDDEN', 'SELECTED', 'CONFIRMED']\n        - **comment** (Type: string):\n        - **componentIdentifier** (Type: object):\n          - **coordinates** (Type: object):\n            - **Additional Properties**:\n              - **property value** (Type: string):\n          - **format** (Type: string):\n        - **id** (Type: string):\n      - **ownerId** (Type: string):\n      - **ownerName** (Type: string):\n      - **ownerType** (Type: string):\n          - Enum: ['application', 'organization', 'repository_container', 'repository_manager', 'repository', 'global']\n"

// NewGetAppliedLicenseOverridesMCPTool creates the MCP Tool instance for GetAppliedLicenseOverrides
func NewGetAppliedLicenseOverridesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAppliedLicenseOverrides",
		"Use this method to retrieve the applied license overrides for a component.\n\nPermissions required: View IQ Elements",
		[]byte(GetAppliedLicenseOverridesInputSchema),
	)
}

// GetAppliedLicenseOverridesHandler is the handler function for the GetAppliedLicenseOverrides tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAppliedLicenseOverridesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/licenseOverrides/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAppliedLicenseOverrides")
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
