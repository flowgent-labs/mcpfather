package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetUserTokenByUsernameAndRealmId tool
const GetUserTokenByUsernameAndRealmIdInputSchema = "{\n  \"properties\": {\n    \"realm\": {\n      \"default\": \"Internal\",\n      \"description\": \"Enter the realmId. Possible values are " + "\x60" + "Internal" + "\x60" + ", " + "\x60" + "SAML" + "\x60" + " , " + "\x60" + "OAUTH2" + "\x60" + " , and " + "\x60" + "Crowd" + "\x60" + ".\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"Enter the username.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"username\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetUserTokenByUsernameAndRealmId tool (Status: 200, Content-Type: application/json)
const GetUserTokenByUsernameAndRealmIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the " + "\x60" + "userCode" + "\x60" + ", " + "\x60" + "username" + "\x60" + " and the name of the IQ server " + "\x60" + "realm" + "\x60" + ".\n\n## Response Structure\n\n- Structure (Type: object):\n  - **lastAccessTime** (Type: string, date-time):\n  - **passCode** (Type: string):\n  - **realm** (Type: string):\n  - **userCode** (Type: string):\n  - **username** (Type: string):\n  - **createTime** (Type: string, date-time):\n  - **expirationDate** (Type: string, date-time):\n"

// NewGetUserTokenByUsernameAndRealmIdMCPTool creates the MCP Tool instance for GetUserTokenByUsernameAndRealmId
func NewGetUserTokenByUsernameAndRealmIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUserTokenByUsernameAndRealmId",
		"Use this method to retrieve a user token by specifying a username and realmId.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetUserTokenByUsernameAndRealmIdInputSchema),
	)
}

// GetUserTokenByUsernameAndRealmIdHandler is the handler function for the GetUserTokenByUsernameAndRealmId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUserTokenByUsernameAndRealmIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/userTokens/{username}", args, []string{"username"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetUserTokenByUsernameAndRealmId")
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
