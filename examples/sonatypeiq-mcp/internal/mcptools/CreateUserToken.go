package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the CreateUserToken tool
const CreateUserTokenInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the CreateUserToken tool (Status: 200, Content-Type: application/json)
const CreateUserTokenResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the generated user token consisting of " + "\x60" + "userCode" + "\x60" + ", " + "\x60" + "username" + "\x60" + " " + "\x60" + "passCode" + "\x60" + ", and the IQ Server " + "\x60" + "realm" + "\x60" + ".\n\n## Response Structure\n\n- Structure (Type: object):\n  - **userCode** (Type: string):\n  - **username** (Type: string):\n  - **createTime** (Type: string, date-time):\n  - **expirationDate** (Type: string, date-time):\n  - **lastAccessTime** (Type: string, date-time):\n  - **passCode** (Type: string):\n  - **realm** (Type: string):\n"

// NewCreateUserTokenMCPTool creates the MCP Tool instance for CreateUserToken
func NewCreateUserTokenMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateUserToken",
		"Use this method to generate a user token for the currently logged in user.\n\nPermissions required: None",
		[]byte(CreateUserTokenInputSchema),
	)
}

// CreateUserTokenHandler is the handler function for the CreateUserToken tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateUserTokenHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/userTokens/currentUser", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "CreateUserToken")
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
