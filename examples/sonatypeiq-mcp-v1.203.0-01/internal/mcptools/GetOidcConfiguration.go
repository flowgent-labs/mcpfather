package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetOidcConfiguration tool
const GetOidcConfigurationInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetOidcConfiguration tool (Status: 200, Content-Type: application/json)
const GetOidcConfigurationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:\n - " + "\x60" + "oidcConfiguration" + "\x60" + " field that contains all the oidc configuration data \n - " + "\x60" + "oAuth2Configuration" + "\x60" + " field that contains the OAuth2 configuration required for oidc\n\n## Response Structure\n\n- Structure (Type: object):\n  - **oidcConfiguration** (Type: object):\n    - **tokenRequestCustomParamsJson** (Type: string):\n    - **authorizationCustomParamsJson** (Type: string):\n    - **clientId** (Type: string):\n    - **clientSecret** (Type: string):\n    - **idpAuthorizationUrl** (Type: string):\n    - **idpIssuer** (Type: string):\n    - **idpTokenUrl** (Type: string):\n  - **oauth2Configuration** (Type: object):\n    - **emailClaim** (Type: string):\n    - **firstNameClaim** (Type: string):\n    - **exactMatchClaimsJson** (Type: string):\n    - **groupsClaim** (Type: string):\n    - **idpJwks** (Type: string):\n    - **idpIssuer** (Type: string):\n    - **idpJwksUrl** (Type: string):\n    - **idpJwsAlgorithm** (Type: string):\n    - **lastNameClaim** (Type: string):\n    - **usernameClaim** (Type: string):\n"

// NewGetOidcConfigurationMCPTool creates the MCP Tool instance for GetOidcConfiguration
func NewGetOidcConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetOidcConfiguration",
		"Use this method to retrieve the OIDC configuration.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetOidcConfigurationInputSchema),
	)
}

// GetOidcConfigurationHandler is the handler function for the GetOidcConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetOidcConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/oidc", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetOidcConfiguration")
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
