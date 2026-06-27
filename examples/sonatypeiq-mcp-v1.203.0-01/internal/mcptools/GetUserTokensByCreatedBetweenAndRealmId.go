package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetUserTokensByCreatedBetweenAndRealmId tool
const GetUserTokensByCreatedBetweenAndRealmIdInputSchema = "{\n  \"properties\": {\n    \"createdAfter\": {\n      \"description\": \"Enter the start date for the date range in " + "\x60" + "yyyy-mm-dd" + "\x60" + " format.\",\n      \"type\": \"string\"\n    },\n    \"createdBefore\": {\n      \"description\": \"Enter the end date for the date range in " + "\x60" + "yyyy-mm-dd" + "\x60" + " format.\",\n      \"type\": \"string\"\n    },\n    \"realm\": {\n      \"default\": \"Internal\",\n      \"description\": \"Enter the " + "\x60" + "realmId" + "\x60" + ". Possible values are " + "\x60" + "Internal" + "\x60" + ", " + "\x60" + "SAML" + "\x60" + " , " + "\x60" + "OAUTH2" + "\x60" + ", and " + "\x60" + "Crowd" + "\x60" + ".\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetUserTokensByCreatedBetweenAndRealmId tool (Status: 200, Content-Type: application/json)
const GetUserTokensByCreatedBetweenAndRealmIdResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a list of user tokens, each containing a " + "\x60" + "userCode" + "\x60" + ", " + "\x60" + "username" + "\x60" + " and the name of the IQ server " + "\x60" + "realm" + "\x60" + ".\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **passCode** (Type: string):\n    - **realm** (Type: string):\n    - **userCode** (Type: string):\n    - **username** (Type: string):\n    - **createTime** (Type: string, date-time):\n    - **expirationDate** (Type: string, date-time):\n    - **lastAccessTime** (Type: string, date-time):\n"

// NewGetUserTokensByCreatedBetweenAndRealmIdMCPTool creates the MCP Tool instance for GetUserTokensByCreatedBetweenAndRealmId
func NewGetUserTokensByCreatedBetweenAndRealmIdMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUserTokensByCreatedBetweenAndRealmId",
		"Use this method to retrieve user tokens created within a date range, in the supported IQ Server realms.\n\nPermissions required: Edit System Configuration and Users.",
		[]byte(GetUserTokensByCreatedBetweenAndRealmIdInputSchema),
	)
}

// GetUserTokensByCreatedBetweenAndRealmIdHandler is the handler function for the GetUserTokensByCreatedBetweenAndRealmId tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUserTokensByCreatedBetweenAndRealmIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/userTokens", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetUserTokensByCreatedBetweenAndRealmId")
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
