package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the CreateUser tool
const CreateUserInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Details of the user to be created\",\n      \"properties\": {\n        \"email\": {\n          \"example\": \"someuser@someemail.com\",\n          \"type\": \"string\"\n        },\n        \"fullName\": {\n          \"example\": \"Some User\",\n          \"type\": \"string\"\n        },\n        \"notifyViaEmail\": {\n          \"example\": true,\n          \"type\": \"boolean\"\n        },\n        \"password\": {\n          \"example\": \"password\",\n          \"type\": \"string\"\n        },\n        \"userName\": {\n          \"example\": \"user1\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the CreateUser tool (Status: 201, Content-Type: application/json)
const CreateUserResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** application/json\n\n> returns a response with generated UserKey for the created user.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateUser tool (Status: 400, Content-Type: application/json)
const CreateUserResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> returned if any error occurs while creating the user\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateUser tool (Status: 401, Content-Type: application/json)
const CreateUserResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> returned if an anonymous (or unauthenticated) user tries to create a user\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateUser tool (Status: 402, Content-Type: application/json)
const CreateUserResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 402\n\n**Content-Type:** application/json\n\n> returned if no more licenses available to create a user\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateUser tool (Status: 403, Content-Type: application/json)
const CreateUserResponseTemplate_E = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> returned if user does not have enough permission to create a user\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateUser tool (Status: 409, Content-Type: application/json)
const CreateUserResponseTemplate_F = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 409\n\n**Content-Type:** application/json\n\n> returned if the user with the same userName already exists\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreateUserMCPTool creates the MCP Tool instance for CreateUser
func NewCreateUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateUser",
		"Create user - One of the following options could be used:\n\n1. Create a user with a specified password. The userName, fullName, email and password needs to be specified.\n\n2. Create a user with an email notification to the user. The userName, fullName, email and notifyViaEmail (true) needs to be specified.\n\n**Requirements**:\n\n- The userName should not be null or blank\n\n- The userName should not contain any of these characters \\ , + < > ' \"\n\n- The userName should not contain any whitespace characters\n\n- The userName should not be \"anonymous\"\n\n- The userName should not contain any upper case characters\n\n- The fullName should not be null or blank\n\n- The fullName should not contain any of these characters\n\n- The fullName should not be \"anonymous\"\n\n- The email should not be null or blank\n\n- The email should be a valid email address\n\n- If notifyViaEmail is false then the password should not be null or blank\n\n- If notifyViaEmail is true then the password should not be specified",
		[]byte(CreateUserInputSchema),
	)
}

// CreateUserHandler is the handler function for the CreateUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/confluence/rest/api/admin/user", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateUser"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
