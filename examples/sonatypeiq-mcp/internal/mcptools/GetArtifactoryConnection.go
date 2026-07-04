package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetArtifactoryConnection tool
const GetArtifactoryConnectionInputSchema = "{\n  \"properties\": {\n    \"artifactoryConnectionId\": {\n      \"description\": \"Enter the Artifactory connection ID.\",\n      \"type\": \"string\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Enter the internal ID of the owner.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the owner type.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"artifactoryConnectionId\",\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetArtifactoryConnection tool (Status: 200, Content-Type: application/json)
const GetArtifactoryConnectionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the details of the requested Artifactory connection.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **username** (Type: string):\n  - **artifactoryConnectionId** (Type: string):\n  - **baseUrl** (Type: string):\n  - **isAnonymous** (Type: boolean):\n  - **ownerId** (Type: string):\n  - **ownerType** (Type: string):\n      - Enum: ['application', 'organization', 'repository_container', 'repository_manager', 'repository', 'global']\n  - **password** (Type: string):\n"

// NewGetArtifactoryConnectionMCPTool creates the MCP Tool instance for GetArtifactoryConnection
func NewGetArtifactoryConnectionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetArtifactoryConnection",
		"Use this method to retrieve details for an Artifactory connection.\n\nPermissions required: View IQ Elements",
		[]byte(GetArtifactoryConnectionInputSchema),
	)
}

// GetArtifactoryConnectionHandler is the handler function for the GetArtifactoryConnection tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetArtifactoryConnectionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/artifactoryConnection/{ownerType}/{internalOwnerId}/{artifactoryConnectionId}", args, []string{"artifactoryConnectionId", "internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetArtifactoryConnection")
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
