package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the AddApplication tool
const AddApplicationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify the applicationId, application name and the organizationId under which the application should be created. " + "\x60" + "contactUserName" + "\x60" + " corresponds to the 'contact' field in the UI and represents the user name. If LDAP is used for authentication, you can use LDAP usernames." + "\x60" + "tagId" + "\x60" + " is the internal identifier for the Application Category that you want to apply to the application. Use the Application Categories REST API for the available categories and the corresponding tagIds.\",\n      \"properties\": {\n        \"applicationTags\": {\n          \"items\": {\n            \"properties\": {\n              \"applicationId\": {\n                \"type\": \"string\"\n              },\n              \"id\": {\n                \"type\": \"string\"\n              },\n              \"tagId\": {\n                \"type\": \"string\"\n              }\n            },\n            \"type\": \"object\"\n          },\n          \"type\": \"array\"\n        },\n        \"contactUserName\": {\n          \"type\": \"string\"\n        },\n        \"id\": {\n          \"type\": \"string\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"organizationId\": {\n          \"type\": \"string\"\n        },\n        \"publicId\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the AddApplication tool (Status: 200, Content-Type: application/json)
const AddApplicationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains application details for the application created using this method.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n  - **organizationId** (Type: string):\n  - **publicId** (Type: string):\n  - **applicationTags** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n      - **tagId** (Type: string):\n      - **applicationId** (Type: string):\n  - **contactUserName** (Type: string):\n  - **id** (Type: string):\n"

// NewAddApplicationMCPTool creates the MCP Tool instance for AddApplication
func NewAddApplicationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AddApplication",
		"Use this method to create an application under an organization. Use the Organization REST API to obtain organizationId.\n\nPermissions required: Add Application (on parent organization)",
		[]byte(AddApplicationInputSchema),
	)
}

// AddApplicationHandler is the handler function for the AddApplication tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddApplicationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/applications", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AddApplication")
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
