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

// Input Schema for the StoreTemporaryAvatarUsingMultiPart1 tool
const StoreTemporaryAvatarUsingMultiPart1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"contentType\": {\n          \"type\": \"string\"\n        },\n        \"formField\": {\n          \"type\": \"boolean\"\n        },\n        \"inputStream\": {\n          \"type\": \"object\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"size\": {\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"value\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"Project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the StoreTemporaryAvatarUsingMultiPart1 tool (Status: 201, Content-Type: text/html)
const StoreTemporaryAvatarUsingMultiPart1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 201\n\n**Content-Type:** text/html\n\n> Temporary avatar cropping instructions embeded in HTML page. Error messages will also be embeded in the page.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **url** (Type: string):\n      - Example: 'http://example.com/jira/secure/temporaryavatar?cropped=true'\n  - **cropperOffsetX** (Type: integer, int32):\n      - Example: '50'\n  - **cropperOffsetY** (Type: integer, int32):\n      - Example: '50'\n  - **cropperWidth** (Type: integer, int32):\n      - Example: '120'\n  - **needsCropping** (Type: boolean):\n      - Example: 'true'\n"

// NewStoreTemporaryAvatarUsingMultiPart1MCPTool creates the MCP Tool instance for StoreTemporaryAvatarUsingMultiPart1
func NewStoreTemporaryAvatarUsingMultiPart1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"StoreTemporaryAvatarUsingMultiPart1",
		"Store temporary avatar using multipart - Creates temporary avatar using multipart. The response is sent back as JSON stored in a textarea. This is because\nthe client uses remote iframing to submit avatars using multipart. So we must send them a valid HTML page back from\nwhich the client parses the JSON.\n",
		[]byte(StoreTemporaryAvatarUsingMultiPart1InputSchema),
	)
}

// StoreTemporaryAvatarUsingMultiPart1Handler is the handler function for the StoreTemporaryAvatarUsingMultiPart1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func StoreTemporaryAvatarUsingMultiPart1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	upstreamURL := upstream + "/rest/api/2/project/{projectIdOrKey}/avatar/temporary"

	req, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream request: %w", err)
	}
	req.Header.Set("Content-Type", "multipart/form-data")

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

	mcputils.LogRequest("POST", upstreamURL, nil, req.Header, nil)

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

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "StoreTemporaryAvatarUsingMultiPart1")
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
