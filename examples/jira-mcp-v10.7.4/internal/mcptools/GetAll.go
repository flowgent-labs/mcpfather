package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAll tool
const GetAllInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAll tool (Status: 200, Content-Type: application/json)
const GetAllResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns all ApplicationRoles in the system\n\n## Response Structure\n\n- Structure (Type: object):\n  - **groups** (Type: array):\n      - Unique Items: true\n      - Example: '[\"jira-software-users\",\"jira-testers\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-software-users\",\"jira-testers\"]'\n  - **userCount** (Type: integer, int32):\n      - Example: '5'\n  - **defined** (Type: boolean):\n      - Example: 'false'\n  - **hasUnlimitedSeats** (Type: boolean):\n      - Example: 'false'\n  - **numberOfSeats** (Type: integer, int32):\n      - Example: '10'\n  - **platform** (Type: boolean):\n      - Example: 'false'\n  - **remainingSeats** (Type: integer, int32):\n      - Example: '5'\n  - **name** (Type: string):\n      - Example: 'Jira Software'\n  - **selectedByDefault** (Type: boolean):\n      - Example: 'false'\n  - **userCountDescription** (Type: string):\n      - Example: '5 developers'\n  - **defaultGroups** (Type: array):\n      - Unique Items: true\n      - Example: '[\"jira-software-users\"]'\n    - **Items** (Type: string):\n        - Example: '[\"jira-software-users\"]'\n  - **key** (Type: string):\n      - Example: 'jira-software'\n"

// NewGetAllMCPTool creates the MCP Tool instance for GetAll
func NewGetAllMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAll",
		"Get all application roles in the system - Returns all application roles in the system.",
		[]byte(GetAllInputSchema),
	)
}

// GetAllHandler is the handler function for the GetAll tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/applicationrole", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAll"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
