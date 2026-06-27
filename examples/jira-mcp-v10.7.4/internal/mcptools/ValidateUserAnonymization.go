package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the ValidateUserAnonymization tool
const ValidateUserAnonymizationInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Parameter used to include parts of the response.\",\n      \"type\": \"string\"\n    },\n    \"userKey\": {\n      \"description\": \"The key of the user to validate anonymization for.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewValidateUserAnonymizationMCPTool creates the MCP Tool instance for ValidateUserAnonymization
func NewValidateUserAnonymizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ValidateUserAnonymization",
		"Get validation for user anonymization - Validates user anonymization process.",
		[]byte(ValidateUserAnonymizationInputSchema),
	)
}

// ValidateUserAnonymizationHandler is the handler function for the ValidateUserAnonymization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ValidateUserAnonymizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/anonymization", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ValidateUserAnonymization"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
