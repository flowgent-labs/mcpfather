package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the AddUserMappings tool
const AddUserMappingsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"\\u003cul\\u003e\\u003cli\\u003eSpecify the " + "\x60" + "role" + "\x60" + " in lowercase, without whitespaces.\\u003c/li\\u003e\\u003cli\\u003e" + "\x60" + "mappings" + "\x60" + " is an array of objects consisting of " + "\x60" + "from" + "\x60" + " and " + "\x60" + "to" + "\x60" + " fields.\\u003c/li\\u003e\\u003cli\\u003eAllowed values for the " + "\x60" + "from" + "\x60" + " field are " + "\x60" + "SCM_USERNAME" + "\x60" + ", " + "\x60" + "SCM_EMAIL" + "\x60" + ", " + "\x60" + "SCM_FULLNAME" + "\x60" + ", " + "\x60" + "GITLOG_EMAIL" + "\x60" + ", " + "\x60" + "GITLOG_FULLNAME" + "\x60" + ".\\u003c/li\\u003e\\u003cli\\u003eAllowed values for " + "\x60" + "to" + "\x60" + " field are " + "\x60" + "IQ_USERNAME" + "\x60" + ", " + "\x60" + "IQ_EMAIL" + "\x60" + ", " + "\x60" + "IQ_FULLNAME" + "\x60" + ".\\u003c/li\\u003e\\u003cli\\u003eAny combination of " + "\x60" + "from" + "\x60" + " and " + "\x60" + "to" + "\x60" + " fields can be used.\\u003c/li\\u003e\\u003c/ul\\u003e\",\n      \"properties\": {\n        \"mappings\": {\n          \"items\": {\n            \"properties\": {\n              \"from\": {\n                \"enum\": [\n                  \"SCM_USERNAME\",\n                  \"SCM_EMAIL\",\n                  \"SCM_FULLNAME\",\n                  \"GITLOG_EMAIL\",\n                  \"GITLOG_FULLNAME\"\n                ],\n                \"type\": \"string\"\n              },\n              \"to\": {\n                \"enum\": [\n                  \"IQ_USERNAME\",\n                  \"IQ_EMAIL\",\n                  \"IQ_FULLNAME\"\n                ],\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        },\n        \"role\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"organizationId\": {\n      \"description\": \"Enter the organizationId. Use " + "\x60" + "ROOT_ORGANIZATION_ID" + "\x60" + " for the root organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"organizationId\"\n  ],\n  \"type\": \"object\"\n}"

// NewAddUserMappingsMCPTool creates the MCP Tool instance for AddUserMappings
func NewAddUserMappingsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddUserMappings",
		"Use this method to apply user mappings from SCM (GitHub) to Lifecycle. The user mappings will be inherited by all child organizations and applications in the organization hierarchy. If a user mapping for an organization already exists, it will be replaced with new mappings provided here.\n\nPermissions required: Edit IQ Elements",
		[]byte(AddUserMappingsInputSchema),
	)
}

// AddUserMappingsHandler is the handler function for the AddUserMappings tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddUserMappingsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/sourceControl/automaticRoleAssignment/userMappings/{organizationId}", args, []string{"organizationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddUserMappings")
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
