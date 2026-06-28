package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the TestArtifactoryConnection tool
const TestArtifactoryConnectionInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Enter values for the Artifactory connection.\\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "baseUrl" + "\x60" + " is the baseURL of the Artifactory instance.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "username" + "\x60" + " and " + "\x60" + "password" + "\x60" + " to authenticate the Artifactory connection.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"artifactoryConnectionId\": {\n          \"type\": \"string\"\n        },\n        \"baseUrl\": {\n          \"type\": \"string\"\n        },\n        \"isAnonymous\": {\n          \"type\": \"boolean\"\n        },\n        \"ownerId\": {\n          \"type\": \"string\"\n        },\n        \"ownerType\": {\n          \"enum\": [\n            \"application\",\n            \"organization\",\n            \"repository_container\",\n            \"repository_manager\",\n            \"repository\",\n            \"global\"\n          ],\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"internalOwnerId\": {\n      \"description\": \"Enter the internal ID of the owner.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Select the owner type.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the TestArtifactoryConnection tool (Status: 204, Content-Type: application/json)
const TestArtifactoryConnectionResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 204\n\n**Content-Type:** application/json\n\n> The response contains the " + "\x60" + "code" + "\x60" + " and " + "\x60" + "message" + "\x60" + " indicating the status of the connection.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **code** (Type: integer, int32):\n  - **message** (Type: string):\n"

// NewTestArtifactoryConnectionMCPTool creates the MCP Tool instance for TestArtifactoryConnection
func NewTestArtifactoryConnectionMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"TestArtifactoryConnection",
		"Use this method to test an Artifactory connection for the specified owner.\n\nPermissons required: View IQ Elements",
		[]byte(TestArtifactoryConnectionInputSchema),
	)
}

// TestArtifactoryConnectionHandler is the handler function for the TestArtifactoryConnection tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func TestArtifactoryConnectionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "*/*"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/config/artifactoryConnection/{ownerType}/{internalOwnerId}/test", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "TestArtifactoryConnection")
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
