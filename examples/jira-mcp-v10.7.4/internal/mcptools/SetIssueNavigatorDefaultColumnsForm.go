package mcptools

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"net/http"
	"os"
	"strings"
	"time"
)

// Input Schema for the SetIssueNavigatorDefaultColumnsForm tool
const SetIssueNavigatorDefaultColumnsFormInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"columns\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// NewSetIssueNavigatorDefaultColumnsFormMCPTool creates the MCP Tool instance for SetIssueNavigatorDefaultColumnsForm
func NewSetIssueNavigatorDefaultColumnsFormMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetIssueNavigatorDefaultColumnsForm",
		"Set default system columns for issue navigator using form - Sets the default system columns for issue navigator. Admin permission will be required.",
		[]byte(SetIssueNavigatorDefaultColumnsFormInputSchema),
	)
}

// SetIssueNavigatorDefaultColumnsFormHandler is the handler function for the SetIssueNavigatorDefaultColumnsForm tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetIssueNavigatorDefaultColumnsFormHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	// Upload: read local file and send as raw body
	localFilePath := ""
	if fp, ok := args["local_file_path"]; ok {
		if s, ok := fp.(string); ok {
			localFilePath = s
		}
	}
	if localFilePath == "" {
		return mcp.NewToolResultError("missing required argument: local_file_path"), nil
	}

	fileData, err := os.ReadFile(localFilePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file %s: %v", localFilePath, err)), nil
	}

	startTime := time.Now()
	upstreamURL := upstream + "/rest/api/2/settings/columns"

	req, err := http.NewRequestWithContext(ctx, "PUT", upstreamURL, bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if forwarded := mcputils.GetHTTPHeaders(ctx); forwarded != nil {
		for key, values := range forwarded {
			lowerKey := strings.ToLower(key)
			if lowerKey == "host" || lowerKey == "connection" || lowerKey == "keep-alive" || lowerKey == "proxy-authenticate" || lowerKey == "proxy-authorization" || lowerKey == "te" || lowerKey == "trailer" || lowerKey == "transfer-encoding" || lowerKey == "upgrade" || lowerKey == "authorization" || lowerKey == "cookie" || lowerKey == "content-length" {
				continue
			}
			for _, v := range values {
				req.Header.Add(key, v)
			}
		}
	}
	if req.Header.Get("Authorization") == "" {
		if token := mcputils.GetUpstreamToken(); token != "" {
			req.Header.Set("Authorization", mcputils.FormatAuthorizationHeader(token))
		}
	}

	if cookie := mcputils.GetUpstreamCookie(); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	mcputils.LogRequest("PUT", upstreamURL, nil, req.Header, nil)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "SetIssueNavigatorDefaultColumnsForm"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
