package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetSamlConfiguration tool
const GetSamlConfigurationInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetSamlConfiguration tool (Status: 200, Content-Type: application/json)
const GetSamlConfigurationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains:<ul><li>" + "\x60" + "identityProviderName" + "\x60" + " the name of the Identity Provider that is displayed on the login page when SAML is configured.</li><li>" + "\x60" + "entityId" + "\x60" + " is the URI that IQ Server uses to identify itself in requests to the SSOservice.</li><li>" + "\x60" + "firstNameAttribute" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider and uses as the user's first name.</li><li>" + "\x60" + "lastNameAttribute" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider and uses as the user's last name.</li><li>" + "\x60" + "emailAttributeName" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider to determine the user's email address.</li><li>" + "\x60" + "usernameAttributeName" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider to determine the username or id.</li><li>" + "\x60" + "groupAttributeName" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider to determine the groups the user belongs to.</li><li>" + "\x60" + "validateResponseSignature" + "\x60" + " indicates whether the SAML responses from the identity provider  are cryptographically signed. A " + "\x60" + "null" + "\x60" + " value indicates that this setting is derived from the SAML metadata from the identity provider performing signature validation if a signing key (" + "\x60" + "KeyDescriptor" + "\x60" + ") is included.<li>" + "\x60" + "validateAssertionSignature" + "\x60" + " indicates whether the SAML assertions from the identity provider  are cryptographically signed. A " + "\x60" + "null" + "\x60" + " value indicates that this setting is derived from  the SAML metadata from the identity provider performing signature validation if a signing key (" + "\x60" + "KeyDescriptor" + "\x60" + ") is included.</li><li>" + "\x60" + "identityProviderMetadataXml" + "\x60" + " is the metadata of the identity provider.</li></ul>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **identityProviderName** (Type: string):\n  - **groupsAttributeName** (Type: string):\n  - **identityProviderMetadataXml** (Type: string):\n  - **usernameAttributeName** (Type: string):\n  - **validateResponseSignature** (Type: boolean):\n  - **entityId** (Type: string):\n  - **lastNameAttributeName** (Type: string):\n  - **validateAssertionSignature** (Type: boolean):\n  - **emailAttributeName** (Type: string):\n  - **firstNameAttributeName** (Type: string):\n"

// NewGetSamlConfigurationMCPTool creates the MCP Tool instance for GetSamlConfiguration
func NewGetSamlConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSamlConfiguration",
		"Use this method to inspect the SAML configuration.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(GetSamlConfigurationInputSchema),
	)
}

// GetSamlConfigurationHandler is the handler function for the GetSamlConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSamlConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/config/saml", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSamlConfiguration")
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
