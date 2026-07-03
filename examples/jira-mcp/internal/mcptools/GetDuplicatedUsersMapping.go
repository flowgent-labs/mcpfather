package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetDuplicatedUsersMapping tool
const GetDuplicatedUsersMappingInputSchema = "{\n  \"properties\": {\n    \"flush\": {\n      \"description\": \"if set to true forces cache flush, user must be sysadmin for this parameter to have an effect.\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetDuplicatedUsersMapping tool (Status: 200, Content-Type: application/json)
const GetDuplicatedUsersMappingResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns all avatars which are visible for the currently logged in user.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **id** (Type: string):\n      - Example: '1000'\n  - **owner** (Type: string):\n      - Example: 'fred'\n  - **selected** (Type: boolean):\n"

// NewGetDuplicatedUsersMappingMCPTool creates the MCP Tool instance for GetDuplicatedUsersMapping
func NewGetDuplicatedUsersMappingMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetDuplicatedUsersMapping",
		"Get duplicated users mapping - Returns duplicated users mapped to their directories with an indication if their accounts are active or not.\nDuplicated means that the user has an account in more than one directory and either more than one account is active\nor the only active account does not belong to the directory with the highest priority.\nThe data returned by this endpoint is cached for 10 minutes and the cache is flushed when any User Directory\nis added, removed, enabled, disabled, or synchronized.\nA System Administrator can also flush the cache manually.\nRelated JAC ticket: https://jira.atlassian.com/browse/JRASERVER-68797",
		[]byte(GetDuplicatedUsersMappingInputSchema),
	)
}

// GetDuplicatedUsersMappingHandler is the handler function for the GetDuplicatedUsersMapping tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetDuplicatedUsersMappingHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/duplicated/list", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetDuplicatedUsersMapping")
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
