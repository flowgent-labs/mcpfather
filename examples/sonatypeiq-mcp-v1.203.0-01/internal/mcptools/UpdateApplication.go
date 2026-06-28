package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the UpdateApplication tool
const UpdateApplicationInputSchema = "{\n  \"properties\": {\n    \"applicationId\": {\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"Specify the applicationId, application name and the organizationId under which  the application exists. " + "\x60" + "contactUserName" + "\x60" + " corresponds to the 'contact' field in the UI and represents the user name. If LDAP is used for authentication, you can use LDAP usernames." + "\x60" + "tagId" + "\x60" + " is the internal identifier for the Application Category that you want to apply to the application. . Use the Application Categories REST API for the available categories and the corresponding tagIds.\",\n      \"properties\": {\n        \"applicationTags\": {\n          \"items\": {\n            \"properties\": {\n              \"applicationId\": {\n                \"type\": \"string\"\n              },\n              \"id\": {\n                \"type\": \"string\"\n              },\n              \"tagId\": {\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        },\n        \"contactUserName\": {\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"organizationId\": {\n          \"type\": \"string\"\n        },\n        \"publicId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"required\": [\n    \"applicationId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateApplication tool (Status: 200, Content-Type: application/json)
const UpdateApplicationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the updated application name, contact user name and application tags,  for the applicationId provided\n\n## Response Structure\n\n- Structure (Type: object):\n  - **applicationTags** (Type: array):\n    - **Items** (Type: object):\n      - **applicationId** (Type: string):\n      - **id** (Type: string):\n      - **tagId** (Type: string):\n  - **contactUserName** (Type: string):\n  - **id** (Type: string):\n  - **name** (Type: string):\n  - **organizationId** (Type: string):\n  - **publicId** (Type: string):\n"

// NewUpdateApplicationMCPTool creates the MCP Tool instance for UpdateApplication
func NewUpdateApplicationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateApplication",
		"Use this method to update the application name, application tags or the contact user name for an existing application by providing the applicationId. \n\nNOTE: This method cannot be used to change the organizationId of an application.\n\nPermissions required: Edit IQ Elements",
		[]byte(UpdateApplicationInputSchema),
	)
}

// UpdateApplicationHandler is the handler function for the UpdateApplication tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateApplicationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/applications/{applicationId}", args, []string{"applicationId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "UpdateApplication")
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
