package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetOwnerArtifactoryConnection tool
const GetOwnerArtifactoryConnectionInputSchema = "{\n  \"properties\": {\n    \"inherit\": {\n      \"default\": false,\n      \"description\": \"Specify whether to include details from an inherited Artifactory connection.\",\n      \"type\": \"boolean\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Enter the internal ID of the owner.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the owner type.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetOwnerArtifactoryConnection tool (Status: 200, Content-Type: application/json)
const GetOwnerArtifactoryConnectionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of the Artifactory connection.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **ownerDTO** (Type: object):\n    - **ownerId** (Type: string):\n    - **ownerName** (Type: string):\n    - **ownerPublicId** (Type: string):\n    - **ownerType** (Type: string):\n  - **artifactoryConnection** (Type: object):\n    - **password** (Type: string):\n    - **username** (Type: string):\n    - **artifactoryConnectionId** (Type: string):\n    - **baseUrl** (Type: string):\n    - **isAnonymous** (Type: boolean):\n    - **ownerId** (Type: string):\n    - **ownerType** (Type: string):\n        - Enum: ['application', 'organization', 'repository_container', 'repository_manager', 'repository', 'global']\n  - **artifactoryConnectionStatus** (Type: object):\n    - **inheritedFromOrganizationId** (Type: string):\n    - **inheritedFromOrganizationName** (Type: string):\n    - **allowChange** (Type: boolean):\n    - **allowOverride** (Type: boolean):\n    - **enabled** (Type: boolean):\n    - **inheritedFromOrgEnabled** (Type: boolean):\n"

// NewGetOwnerArtifactoryConnectionMCPTool creates the MCP Tool instance for GetOwnerArtifactoryConnection
func NewGetOwnerArtifactoryConnectionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetOwnerArtifactoryConnection",
		"Use this method to retrieve Artifactory connection details by specifying the owner Id.\n\nPermissions required: View IQ Elements",
		[]byte(GetOwnerArtifactoryConnectionInputSchema),
	)
}

// GetOwnerArtifactoryConnectionHandler is the handler function for the GetOwnerArtifactoryConnection tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetOwnerArtifactoryConnectionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/artifactoryConnection/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetOwnerArtifactoryConnection")
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
