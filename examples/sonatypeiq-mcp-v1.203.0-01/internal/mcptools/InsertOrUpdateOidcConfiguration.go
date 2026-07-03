package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the InsertOrUpdateOidcConfiguration tool
const InsertOrUpdateOidcConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"oauth2Configuration\": {\n          \"properties\": {\n            \"emailClaim\": {\n              \"type\": \"string\"\n            },\n            \"exactMatchClaimsJson\": {\n              \"type\": \"string\"\n            },\n            \"firstNameClaim\": {\n              \"type\": \"string\"\n            },\n            \"groupsClaim\": {\n              \"type\": \"string\"\n            },\n            \"idpIssuer\": {\n              \"type\": \"string\"\n            },\n            \"idpJwks\": {\n              \"type\": \"string\"\n            },\n            \"idpJwksUrl\": {\n              \"type\": \"string\"\n            },\n            \"idpJwsAlgorithm\": {\n              \"type\": \"string\"\n            },\n            \"lastNameClaim\": {\n              \"type\": \"string\"\n            },\n            \"usernameClaim\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"oidcConfiguration\": {\n          \"properties\": {\n            \"authorizationCustomParamsJson\": {\n              \"type\": \"string\"\n            },\n            \"clientId\": {\n              \"type\": \"string\"\n            },\n            \"clientSecret\": {\n              \"type\": \"string\"\n            },\n            \"idpAuthorizationUrl\": {\n              \"type\": \"string\"\n            },\n            \"idpIssuer\": {\n              \"type\": \"string\"\n            },\n            \"idpTokenUrl\": {\n              \"type\": \"string\"\n            },\n            \"tokenRequestCustomParamsJson\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewInsertOrUpdateOidcConfigurationMCPTool creates the MCP Tool instance for InsertOrUpdateOidcConfiguration
func NewInsertOrUpdateOidcConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"InsertOrUpdateOidcConfiguration",
		"Use this method to enable SSO using OpenID Connect (OIDC). This request uses the content type application/json to transmit the configuration to IQ Server.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(InsertOrUpdateOidcConfigurationInputSchema),
	)
}

// InsertOrUpdateOidcConfigurationHandler is the handler function for the InsertOrUpdateOidcConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func InsertOrUpdateOidcConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/oidc", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "InsertOrUpdateOidcConfiguration")
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
