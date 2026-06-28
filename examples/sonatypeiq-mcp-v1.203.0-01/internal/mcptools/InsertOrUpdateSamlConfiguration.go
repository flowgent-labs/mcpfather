package mcptools

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"net/http"
	"os"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"strings"
	"time"
)

// Input Schema for the InsertOrUpdateSamlConfiguration tool
const InsertOrUpdateSamlConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"identityProviderXml\": {\n          \"description\": \"Enter the SAML metadata XML of your IdP. Refer to the IdP documentation to obtain this metadata.\",\n          \"format\": \"binary\",\n          \"type\": \"string\"\n        },\n        \"samlConfiguration\": {\n          \"description\": \"Enter the SAML configuration\\u003cul\\u003e\\u003cli\\u003e" + "\x60" + "identityProviderName" + "\x60" + " the name of the Identity Provider that is displayed on the login page when SAML is configured.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "entityId" + "\x60" + " is the URI that IQ Server uses to identify itself in requests to the SSOservice.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "firstNameAttribute" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider and uses as the user's first name.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "lastNameAttribute" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider and uses as the user's last name.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "emailAttributeName" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider to determine the user's email address.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "usernameAttributeName" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider to determine the username or id.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "groupAttributeName" + "\x60" + " is the SAML attribute that IQ Server extracts from the login response of the identity provider to determine the groups the user belongs to.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "validateResponseSignature" + "\x60" + " indicates whether the SAML responses from the identity provider  are cryptographically signed. A " + "\x60" + "null" + "\x60" + " value indicates that this setting is derived from the SAML metadata from the identity provider performing signature validation if a signing key (" + "\x60" + "KeyDescriptor" + "\x60" + ") is included.\\u003cli\\u003e" + "\x60" + "validateAssertionSignature" + "\x60" + " indicates whether the SAML assertions from the identity provider  are cryptographically signed. A " + "\x60" + "null" + "\x60" + " value indicates that this setting is derived from  the SAML metadata from the identity provider performing signature validation if a signing key (" + "\x60" + "KeyDescriptor" + "\x60" + ") is included.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "identityProviderMetadataXml" + "\x60" + " is the metadata of the identity provider.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n          \"properties\": {\n            \"emailAttributeName\": {\n              \"type\": \"string\"\n            },\n            \"entityId\": {\n              \"type\": \"string\"\n            },\n            \"firstNameAttributeName\": {\n              \"type\": \"string\"\n            },\n            \"groupsAttributeName\": {\n              \"type\": \"string\"\n            },\n            \"identityProviderName\": {\n              \"type\": \"string\"\n            },\n            \"lastNameAttributeName\": {\n              \"type\": \"string\"\n            },\n            \"usernameAttributeName\": {\n              \"type\": \"string\"\n            },\n            \"validateAssertionSignature\": {\n              \"type\": \"boolean\"\n            },\n            \"validateResponseSignature\": {\n              \"type\": \"boolean\"\n            }\n          },\n          \"type\": \"object\"\n        }\n      },\n      \"required\": [\n        \"identityProviderXml\",\n        \"samlConfiguration\"\n      ],\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewInsertOrUpdateSamlConfigurationMCPTool creates the MCP Tool instance for InsertOrUpdateSamlConfiguration
func NewInsertOrUpdateSamlConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"InsertOrUpdateSamlConfiguration",
		"Use this method to enable SSO using SAML. This request uses the content type multipart/form-data to transmit the configuration to IQ Server.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(InsertOrUpdateSamlConfigurationInputSchema),
	)
}

// InsertOrUpdateSamlConfigurationHandler is the handler function for the InsertOrUpdateSamlConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func InsertOrUpdateSamlConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	// Upload: read local file and send as raw body
	localFilePath := ""
	if fp, ok := args["local_file_path"]; ok {
		if s, ok := fp.(string); ok {
			localFilePath = s
		}
	}
	if localFilePath == "" {
		return mcp.NewToolResultError("missing required argument: local_file_path"), nil
	}

	fileData, err := os.ReadFile(localFilePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file %s: %v", localFilePath, err)), nil
	}

	startTime := time.Now()
	upstreamURL := upstream + "/api/v2/config/saml"

	req, err := http.NewRequestWithContext(ctx, "PUT", upstreamURL, bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream request: %w", err)
	}
	req.Header.Set("Content-Type", "multipart/form-data")

	if forwarded := mcputils.GetHTTPHeaders(ctx); forwarded != nil {
		for key, values := range forwarded {
			lowerKey := strings.ToLower(key)
			if lowerKey == "host" || lowerKey == "connection" || lowerKey == "keep-alive" || lowerKey == "proxy-authenticate" || lowerKey == "proxy-authorization" || lowerKey == "te" || lowerKey == "trailer" || lowerKey == "transfer-encoding" || lowerKey == "upgrade" || lowerKey == "authorization" || lowerKey == "cookie" || lowerKey == "content-length" || lowerKey == "mcp-session-id" || lowerKey == "content-type" {
				continue
			}
			for _, v := range values {
				req.Header.Add(key, v)
			}
		}
	}
	if req.Header.Get("Authorization") == "" {
		if token := mcputils.GetUpstreamToken(); token != "" {
			req.Header.Set("Authorization", mcputils.FormatAuthorizationHeader(token))
		}
	}

	if cookie := mcputils.GetUpstreamCookie(); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	if mcputils.GetUpstreamConfig().EnableMCPSessionInForwarding {
		if sid := mcputils.GetSessionID(ctx); sid != "" {
			req.Header.Set("X-MCP-Session-ID", sid)
		}
	}

	mcputils.LogRequest("PUT", upstreamURL, nil, req.Header, nil)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "InsertOrUpdateSamlConfiguration")
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
