package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetUserMappingsByOwner tool
const GetUserMappingsByOwnerInputSchema = "{\n  \"properties\": {\n    \"internalOwnerId\": {\n      \"description\": \"Enter the value for internal ownerId.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the value for ownerType.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetUserMappingsByOwner tool (Status: 200, Content-Type: application/json)
const GetUserMappingsByOwnerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "ownerInternalId" + "\x60" + " indicates the owner id for which the user mappings were created.</li><li>" + "\x60" + "inherited" + "\x60" + " is always " + "\x60" + "true" + "\x60" + " if the ownerType is application</li><li>" + "\x60" + "userMapping" + "\x60" + " is an object containing " + "\x60" + "role" + "\x60" + " and " + "\x60" + "mappings" + "\x60" + ".<ul><li> " + "\x60" + "role" + "\x60" + " indicates the role assigned to users during automatic role assignment.</li><li>" + "\x60" + "mappings" + "\x60" + " contain all existing user mappings from the SCM sytem to IQ.</li></ul></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **userMapping** (Type: object):\n    - **mappings** (Type: array):\n      - **Items** (Type: object):\n        - **from** (Type: string):\n            - Enum: ['SCM_USERNAME', 'SCM_EMAIL', 'SCM_FULLNAME', 'GITLOG_EMAIL', 'GITLOG_FULLNAME']\n        - **to** (Type: string):\n            - Enum: ['IQ_USERNAME', 'IQ_EMAIL', 'IQ_FULLNAME']\n    - **role** (Type: string):\n  - **inherited** (Type: boolean):\n  - **ownerInternalId** (Type: string):\n"

// NewGetUserMappingsByOwnerMCPTool creates the MCP Tool instance for GetUserMappingsByOwner
func NewGetUserMappingsByOwnerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUserMappingsByOwner",
		"Use this method to retrieve SCM user mappings for an organization or application.\n\nPermissions required: View IQ Elements",
		[]byte(GetUserMappingsByOwnerInputSchema),
	)
}

// GetUserMappingsByOwnerHandler is the handler function for the GetUserMappingsByOwner tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUserMappingsByOwnerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/sourceControl/automaticRoleAssignment/userMappings/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetUserMappingsByOwner")
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
