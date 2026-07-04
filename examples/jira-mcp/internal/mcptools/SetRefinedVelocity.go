package mcptools

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"net/http"
	"os"
	"strings"
	"time"
)

// Input Schema for the SetRefinedVelocity tool
const SetRefinedVelocityInputSchema = "{\n  \"properties\": {\n    \"boardId\": {\n      \"description\": \"The id of the board on which the property will be set.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"body\": {\n      \"description\": \"The request containing value of the board's property. The value has to a valid, non-empty JSON conforming to http://tools.ietf.org/html/rfc4627. The maximum length of the property value is 32768 bytes.\",\n      \"oneOf\": [\n        {\n          \"title\": \"Schema for application/json\"\n        },\n        {\n          \"properties\": {\n            \"value\": {\n              \"example\": true,\n              \"type\": \"boolean\"\n            }\n          },\n          \"title\": \"Schema for application/x-www-form-urlencoded\",\n          \"type\": \"object\"\n        }\n      ]\n    }\n  },\n  \"required\": [\n    \"boardId\",\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetRefinedVelocityMCPTool creates the MCP Tool instance for SetRefinedVelocity
func NewSetRefinedVelocityMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetRefinedVelocity",
		"Update the board's refined velocity setting - Sets the value of the specified board's refined velocity setting.",
		[]byte(SetRefinedVelocityInputSchema),
	)
}

// SetRefinedVelocityHandler is the handler function for the SetRefinedVelocity tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetRefinedVelocityHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	upstreamURL := upstream + "/rest/agile/1.0/board/{boardId}/settings/refined-velocity"

	req, err := http.NewRequestWithContext(ctx, "PUT", upstreamURL, bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if forwarded := mcputils.GetHTTPHeaders(ctx); forwarded != nil {
		for key, values := range forwarded {
			lowerKey := strings.ToLower(key)
			if lowerKey == "host" || lowerKey == "connection" || lowerKey == "keep-alive" || lowerKey == "proxy-authenticate" || lowerKey == "proxy-authorization" || lowerKey == "te" || lowerKey == "trailer" || lowerKey == "transfer-encoding" || lowerKey == "upgrade" || lowerKey == "authorization" || lowerKey == "cookie" || lowerKey == "content-length" || lowerKey == "mcp-session-id" || lowerKey == "content-type" {
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

	// Always forward MCP session ID as a standard HTTP header.
	// The raw "Mcp-Session-Id"/"mcp-session-id" header from the MCP client is
	// never forwarded as-is because some upstream APIs (e.g. Sonatype IQ)
	// reject non-standard headers with HTTP 400.
	if sid := mcputils.GetSessionID(ctx); sid != "" {
		req.Header.Set("X-MCP-Session-ID", sid)
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

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetRefinedVelocity")
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
