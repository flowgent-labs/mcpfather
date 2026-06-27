package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the Add tool
const AddInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Specify the user details for the new user to be created. All fields except " + "\x60" + "realm" + "\x60" + " are required.\",\n      \"properties\": {\n        \"email\": {\n          \"type\": \"string\"\n        },\n        \"firstName\": {\n          \"type\": \"string\"\n        },\n        \"lastName\": {\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"type\": \"string\"\n        },\n        \"realm\": {\n          \"type\": \"string\"\n        },\n        \"username\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewAddMCPTool creates the MCP Tool instance for Add
func NewAddMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Add",
		"Use this method to create a new user.\n\nPermissions required: Edit System Configuration and Users",
		[]byte(AddInputSchema),
	)
}

// AddHandler is the handler function for the Add tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AddHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/users", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Add")
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
