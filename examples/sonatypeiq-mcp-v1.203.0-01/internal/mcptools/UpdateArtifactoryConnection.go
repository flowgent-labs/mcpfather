package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the UpdateArtifactoryConnection tool
const UpdateArtifactoryConnectionInputSchema = "{\n  \"properties\": {\n    \"artifactoryConnectionId\": {\n      \"description\": \"Enter the Artifactory connection ID.\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"Enter values for the new Artifactory connection.\\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "isAnonymous" + "\x60" + " indicates if the connection is anonymous.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "baseUrl" + "\x60" + " is the baseURL of the Artifactory instance.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "username" + "\x60" + " and " + "\x60" + "password" + "\x60" + " to authenticate the Artifactory connection.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"artifactoryConnectionId\": {\n          \"type\": \"string\"\n        },\n        \"baseUrl\": {\n          \"type\": \"string\"\n        },\n        \"isAnonymous\": {\n          \"type\": \"boolean\"\n        },\n        \"ownerId\": {\n          \"type\": \"string\"\n        },\n        \"ownerType\": {\n          \"enum\": [\n            \"application\",\n            \"organization\",\n            \"repository_container\",\n            \"repository_manager\",\n            \"repository\",\n            \"global\"\n          ],\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Enter the internal ID of the owner.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Specify the owner type.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"artifactoryConnectionId\",\n    \"body\",\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateArtifactoryConnection tool (Status: 200, Content-Type: application/json)
const UpdateArtifactoryConnectionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the updated Artifactory connection details.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **artifactoryConnectionId** (Type: string):\n  - **baseUrl** (Type: string):\n  - **isAnonymous** (Type: boolean):\n  - **ownerId** (Type: string):\n  - **ownerType** (Type: string):\n      - Enum: ['application', 'organization', 'repository_container', 'repository_manager', 'repository', 'global']\n  - **password** (Type: string):\n  - **username** (Type: string):\n"

// NewUpdateArtifactoryConnectionMCPTool creates the MCP Tool instance for UpdateArtifactoryConnection
func NewUpdateArtifactoryConnectionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateArtifactoryConnection",
		"Use this method to update an existing Artifactory connection.\n\nPermissions required: Edit IQ Elements",
		[]byte(UpdateArtifactoryConnectionInputSchema),
	)
}

// UpdateArtifactoryConnectionHandler is the handler function for the UpdateArtifactoryConnection tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateArtifactoryConnectionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/artifactoryConnection/{ownerType}/{internalOwnerId}/{artifactoryConnectionId}", args, []string{"artifactoryConnectionId", "internalOwnerId", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateArtifactoryConnection")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
